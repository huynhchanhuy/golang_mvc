package main

import (
	"github.com/heroku/huhu-backend-app/app"
	"github.com/heroku/huhu-backend-app/config"
	_ "github.com/joho/godotenv/autoload"
	"os"
)

func main() {
	config := config.GetConfig()

	app := &app.App{}
	app.Initialize(config)
	println("Listen on port " + os.Getenv("PORT"))
	app.Run(":" + os.Getenv("PORT"))
}
