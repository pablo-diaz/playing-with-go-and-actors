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
	router.RunHttpServerWithRoutes(getEnvVar("ON_HOST", "localhost"), getEnvVar("ON_PORT", "8080"), routes)
}

func loadEnvVars() {
	err := godotenv.Load(".env.local", ".env")
	if err != nil && !os.IsNotExist(err) {
		panic(fmt.Sprintf("Error loading .env files: %v", err))
	}
}

func getEnvVar(withName string, defaultValueWhenNotPresent string) string {
	maybeValueFoundForEnvVar, exists := os.LookupEnv(withName)
	if !exists {
		return defaultValueWhenNotPresent
	}

	return maybeValueFoundForEnvVar
}
