package main

import (
	"github.com/joho/godotenv"
	"github.com/sstehniy/sendiz.app/server"
)

func main() {
	godotenv.Load()
	db := configureDB()
	server.StartServer(db)
}
