package server

import (
	"database/sql"
)

func StartServer(db *sql.DB) {
	setupApi(db)
}
