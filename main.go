package main

import (
	"fmt"
	"os"

	"example.com/web-service-gin/actorModel"
	"example.com/web-service-gin/router"
	"github.com/joho/godotenv"
)

func main() {
	loadEnvVars()
	routes := router.SetupRouter(actorModel.CreateNewAlbumManager())
	router.RunHttpServerWithRoutes("localhost", "8080", routes)
}

func loadEnvVars() {
	err := godotenv.Load(".env.local", ".env")
	if err != nil && !os.IsNotExist(err) {
		panic(fmt.Sprintf("Error loading .env files: %v", err))
	}
}
