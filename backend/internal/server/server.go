package server

import "github.com/gin-gonic/gin"

func Run(listen string) error {
	router := gin.Default()
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})
	return router.Run(listen)
}
