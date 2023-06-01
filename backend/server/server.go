package server

import (
	"database/sql"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func StartServer(db *sql.DB) {
	router := gin.Default()
	defer router.Run("0.0.0.0:8080")

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000"}
	router.Use(cors.New(config))
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	setupApi(router, db)
	setupWebSocket(router, db)
}
