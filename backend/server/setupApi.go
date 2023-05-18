package server

import (
	"database/sql"

	"github.com/gin-gonic/gin"
)

func setupApi(router *gin.Engine, db *sql.DB) {

	v1 := router.Group("/api/v1")
	addUserRoutes(v1, db)

}
