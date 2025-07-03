package router

import (
	"fmt"
	"os"

	"example.com/web-service-gin/actorModel"
	"example.com/web-service-gin/dbTasks"
	"example.com/web-service-gin/handlers"
	"github.com/gin-gonic/contrib/expvar"
	"github.com/gin-gonic/gin"
)

func SetupRouter(withAlbumManager *actorModel.AlbumManager) *gin.Engine {
	router := gin.Default()

	db, err := dbTasks.ConnectDB(mustGetDbConnectionString())
	if err != nil {
		panic(fmt.Sprintf("Error found when trying to connect to DB: %v", err))
	}

	router.Use(func(c *gin.Context) {
		c.Set("db", db)
		c.Set("albumsManager", withAlbumManager)
		c.Next()
	})

	router.GET("/albums", handlers.GetAlbums)
	router.GET("/albums/:id", handlers.GetAlbumByID)
	router.GET("/albums/new/:id", handlers.GetAlbumById_NewWay)

	router.POST("/albums", handlers.PostAlbums)

	router.GET("/debug/vars", expvar.Handler())

	return router
}

func mustGetDbConnectionString() string {
	result, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		panic("DATABASE_URL environment variable not set")
	}

	return result
}

func RunHttpServerWithRoutes(onHost string, onPort string, withRouter *gin.Engine) {
	withRouter.Run(fmt.Sprintf("%v:%v", onHost, onPort))
}
