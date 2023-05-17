package main

import (
	"github.com/joho/godotenv"
	"github.com/sstehniy/sendiz.app/pkg/api"
)

func main() {
	godotenv.Load()
	db := configureDB()
	api.SetupApi(db)

}
