package actorModel

import (
	"errors"
	"fmt"
	"time"

	"example.com/web-service-gin/models"
	"gorm.io/gorm"
)

const MAX_ENTRIES_FOR_ALBUM_ACTOR_INBOX = 100_000
const SECONDS_TO_STOP_ACTOR = 3

type AlbumId string

type GetAlbumInfoReq struct {
	placeInfoHere chan *ResponseAfterGettingAlbumInfo
}

type AlbumActor struct {
	info  *models.Album
	inbox chan GetAlbumInfoReq
}

type ResponseAfterGettingAlbumInfo struct {
	MaybeAlbumInfoFound *models.Album
	MaybeErrorFound     error
}

func createAlbumActor(albumId AlbumId, usingDb *gorm.DB, chanToNotifyWhenActorStopped chan<- any) (*AlbumActor, error) {
	maybeAlbumFound, err := loadInfoFromPersistence(albumId, usingDb)
	if err != nil {
		return nil, err
	}

	inbox := make(chan GetAlbumInfoReq, MAX_ENTRIES_FOR_ALBUM_ACTOR_INBOX)

	go receiveAllMessages(inbox, maybeAlbumFound, SECONDS_TO_STOP_ACTOR*time.Second, albumId, chanToNotifyWhenActorStopped)

	return &AlbumActor{info: maybeAlbumFound, inbox: inbox}, nil
}

func receiveAllMessages(fromInbox chan GetAlbumInfoReq, withAlbum *models.Album, andFinishAfter time.Duration, albumId AlbumId, chanToNotifyWhenStopped chan<- any) {
	timeout := time.After(andFinishAfter)

	defer func() {
		fmt.Printf("--------- [Album %v] Stopping actor \n", albumId)
		chanToNotifyWhenStopped <- albumId
	}()

loopToProcessAllMessagesFromInbox:
	for {
		select {
		case message := <-fromInbox:
			processRequestToGetAlbumInfo(message, withAlbum)
		case <-timeout:
			break loopToProcessAllMessagesFromInbox // it's time to stop this never ending loop
		}
	}
}

func (a *AlbumActor) placeRequestToGetAlbumInfo(r GetAlbumInfoReq) {
	a.inbox <- r
}

func processRequestToGetAlbumInfo(message GetAlbumInfoReq, infoToReturn *models.Album) {
	message.placeInfoHere <- &ResponseAfterGettingAlbumInfo{MaybeAlbumInfoFound: infoToReturn, MaybeErrorFound: nil}
}

func loadInfoFromPersistence(idToLoad AlbumId, usingDb *gorm.DB) (*models.Album, error) {
	fmt.Printf("--------- [Album %v] Loading album from persistence \n", idToLoad)

	var albumsFoundWithGivenId []models.Album
	result := usingDb.Find(&albumsFoundWithGivenId, "id = ?", string(idToLoad))

	if result.Error != nil {
		return nil, result.Error
	}

	if len(albumsFoundWithGivenId) == 0 {
		return nil, errors.New("album was not found with given Id")
	}

	return &albumsFoundWithGivenId[0], nil
}
