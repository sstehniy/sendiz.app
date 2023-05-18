package server

import (
	"database/sql"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func StartServer(db *sql.DB) {
	router := gin.Default()
	defer router.Run(":8080")

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000"}
	router.Use(cors.New(config))
	setupApi(router, db)
	setupWebSocket(router, db)
}
