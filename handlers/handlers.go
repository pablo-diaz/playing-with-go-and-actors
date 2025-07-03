package handlers

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"example.com/web-service-gin/actorModel"
	"example.com/web-service-gin/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const ALLOW_ONE_RESPONSE_TO_BE_SENT = 1

func GetAlbums(c *gin.Context) {
	db := getDbFromContext(c)
	var albums []models.Album
	result := db.Find(&albums)
	if result.Error != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, albums)
}

func PostAlbums(c *gin.Context) {
	var newAlbum models.Album

	if err := c.BindJSON(&newAlbum); err != nil {
		return
	}

	db := getDbFromContext(c)
	result := db.Create(&newAlbum)
	if result.Error != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	c.IndentedJSON(http.StatusCreated, gin.H{"message": "album created successfully", "album": newAlbum})
}

func GetAlbumByID(c *gin.Context) {
	id := c.Param("id")
	db := getDbFromContext(c)
	var albumsFoundWithGivenId []models.Album
	result := db.Find(&albumsFoundWithGivenId, "id = ?", id)

	if result.Error != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	if len(albumsFoundWithGivenId) == 1 {
		c.IndentedJSON(http.StatusOK, albumsFoundWithGivenId[0])
		return
	}

	if len(albumsFoundWithGivenId) > 1 {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "more than one album found with provided ID"})
		return
	}

	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "album not found"})
}

func GetAlbumById_NewWay(c *gin.Context) {
	id := c.Param("id")
	db := getDbFromContext(c)
	am := getAlbumsManagerFromContext(c)

	responseChan := make(chan *actorModel.ResponseAfterGettingAlbumInfo, ALLOW_ONE_RESPONSE_TO_BE_SENT)

	am.PlaceRequestToGetAlbumInfo(actorModel.RequestToGetAlbumInfo{
		AlbumIdToRequest: actorModel.AlbumId(id),
		UsingDb:          db,
		PlaceInfoHere:    responseChan})

	select {
	case response := <-responseChan:
		if response.MaybeErrorFound != nil {
			if strings.Contains(response.MaybeErrorFound.Error(), "album was not found") {
				c.IndentedJSON(http.StatusNotFound, gin.H{"message": "album not found"})
				return
			}

			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": response.MaybeErrorFound.Error()})
			return
		}

		c.IndentedJSON(http.StatusOK, response.MaybeAlbumInfoFound)
		return

	case <-time.After(10 * time.Second):
		fmt.Fprintf(os.Stderr, "[%v] Request for '%v' timed-out: %v \n", time.Now(), id, c.Request.URL.String())
		c.IndentedJSON(http.StatusGatewayTimeout, gin.H{"error": "stop waiting for response about Album Info"})
		return
	}

}

func getDbFromContext(c *gin.Context) *gorm.DB {
	return c.MustGet("db").(*gorm.DB)
}

func getAlbumsManagerFromContext(c *gin.Context) *actorModel.AlbumManager {
	return c.MustGet("albumsManager").(*actorModel.AlbumManager)
}
