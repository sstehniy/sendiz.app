package main

import (
	"github.com/joho/godotenv"
	"github.com/sstehniy/sendiz.app/server"
)

type UserInitiate struct {
	ID    int64  `json:"id"`
	Phone string `json:"phone"`
}

func main() {
	godotenv.Load()
	db := configureDB()
	defer db.Close()
	server.StartServer(db)
}
