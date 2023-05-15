package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func configureDB() *sql.DB {
	godotenv.Load()
	db, err := sql.Open("mysql", os.Getenv("DSN"))

	if err != nil {
		log.Fatal(err)
	}

	return db
}
