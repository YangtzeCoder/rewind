package server

import (
	"context"

	"rewind/internal/auth"
	"rewind/internal/database"

	"github.com/gin-gonic/gin"
)

func Run(listen, dataDir string) error {
	db, err := database.Open(dataDir)
	if err != nil {
		return err
	}
	defer db.Close()
	authService := auth.NewService(database.NewAuthStore(db))
	if err := authService.Initialize(context.Background()); err != nil {
		return err
	}
	router := gin.New()
	router.Use(gin.Logger(), gin.CustomRecovery(func(c *gin.Context, _ any) {
		if !c.Writer.Written() {
			failure(c, errInternal, "internal server error")
		}
		c.Abort()
	}))
	router.HandleMethodNotAllowed = true
	router.NoRoute(func(c *gin.Context) { failure(c, errNotFound, "not found") })
	router.NoMethod(func(c *gin.Context) { failure(c, errMethodNotAllowed, "method not allowed") })
	api := router.Group("/api/v1")
	registerAuthRoutes(api, authService)
	registerTokenRoutes(api, authService, database.NewTokenStore(db))
	registerProjectRoutes(api, authService, database.NewProjectStore(db))
	return router.Run(listen)
}
