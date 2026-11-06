package server

import (
	"context"
	"net/http"

	"rewind/internal/auth"
	"rewind/internal/database"
	"rewind/internal/storage"

	"github.com/gin-gonic/gin"
)

func Run(listen, dataDir string) error {
	db, err := database.Open(dataDir)
	if err != nil {
		return err
	}
	defer db.Close()
	files, err := storage.NewLocal(dataDir)
	if err != nil {
		return err
	}
	authService := auth.NewService(database.NewAuthStore(db))
	if err := authService.Initialize(context.Background()); err != nil {
		return err
	}
	router := gin.New()
	router.Use(gin.Logger(), gin.CustomRecovery(func(c *gin.Context, recovered any) {
		if recovered == http.ErrAbortHandler {
			panic(http.ErrAbortHandler)
		}
		if !c.Writer.Written() {
			failure(c, errInternal, "internal server error")
		}
		c.Abort()
	}))
	router.HandleMethodNotAllowed = true
	router.NoRoute(func(c *gin.Context) { failure(c, errNotFound, "not found") })
	router.NoMethod(func(c *gin.Context) { failure(c, errMethodNotAllowed, "method not allowed") })
	api := router.Group("/api/v1")
	projects := database.NewProjectStore(db)
	api.GET("/dashboard", (&authHandler{service: authService}).authenticate, func(c *gin.Context) {
		stats, err := projects.Dashboard(c.Request.Context())
		if err != nil {
			internalError(c)
			return
		}
		success(c, gin.H{"project_count": stats.ProjectCount, "file_count": stats.FileCount,
			"version_count": stats.VersionCount, "total_size": stats.TotalSize})
	})
	registerAuthRoutes(api, authService)
	registerTokenRoutes(api, authService, database.NewTokenStore(db))
	registerProjectRoutes(api, authService, projects, files)
	registerBackupRoutes(api, authService, database.NewTokenStore(db), database.NewBackupStore(db), files)
	return router.Run(listen)
}
