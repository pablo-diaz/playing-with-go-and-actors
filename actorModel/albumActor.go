package actorModel

import (
	"errors"
	"fmt"
	"os"
	"time"

	"example.com/web-service-gin/models"
	"example.com/web-service-gin/monitoredChannel"
	"gorm.io/gorm"
)

const MAX_ENTRIES_FOR_ALBUM_ACTOR_INBOX = 1_000
const FREQUENCY_TO_REFRESH_DATA_FROM_PERSISTENCE = 3

type AlbumId string

type GetAlbumInfoReq struct {
	placeInfoHere chan *ResponseAfterGettingAlbumInfo
}

type AlbumActor struct {
	info                                  *models.Album
	lastTimeDataWasFetchedFromPersistence time.Time
	inbox                                 *monitoredChannel.MonitoredChannel[GetAlbumInfoReq]
}

type ResponseAfterGettingAlbumInfo struct {
	MaybeAlbumInfoFound *models.Album
	MaybeErrorFound     error
}

func createAlbumActor(albumId AlbumId, usingDb *gorm.DB) (*AlbumActor, error) {
	maybeAlbumFound, err := loadInfoFromPersistence(albumId, usingDb)
	if err != nil {
		return nil, err
	}

	inbox := monitoredChannel.NewMonitoredChannel[GetAlbumInfoReq](fmt.Sprintf("album_%v_inbox", string(albumId)), MAX_ENTRIES_FOR_ALBUM_ACTOR_INBOX, 1*time.Second)

	newActor := &AlbumActor{
		info:                                  maybeAlbumFound,
		lastTimeDataWasFetchedFromPersistence: time.Now(),
		inbox:                                 inbox,
	}

	go newActor.receiveAllMessages(maybeAlbumFound, albumId, usingDb)

	return newActor, nil
}

func (m *AlbumActor) receiveAllMessages(withAlbum *models.Album, albumId AlbumId, usingDb *gorm.DB) {
	for {
		request := m.inbox.Receive()
		if m.shouldItRefreshDataFromPersistence() {
			m.refreshDataFromPersistence(albumId, usingDb)
		}
		processRequestToGetAlbumInfo(request, withAlbum)
	}
}

func (m *AlbumActor) shouldItRefreshDataFromPersistence() bool {
	return time.Since(m.lastTimeDataWasFetchedFromPersistence).Seconds() >= FREQUENCY_TO_REFRESH_DATA_FROM_PERSISTENCE
}

func (m *AlbumActor) refreshDataFromPersistence(withAlbumId AlbumId, usingDb *gorm.DB) {
	m.lastTimeDataWasFetchedFromPersistence = time.Now()
	maybeAlbumFound, err := loadInfoFromPersistence(withAlbumId, usingDb)

	if err != nil {
		fmt.Fprintf(os.Stderr, "--------- [Album %v] There was an error refreshing data from persistence: %v \n", withAlbumId, err)
		return // we only log this error, but we don't stop it all, so this actor can still return the most recent cached info
	}

	m.info = maybeAlbumFound
	fmt.Printf("--------- [Album %v] Data was refreshed successfully from persistence \n", withAlbumId)
}

func (a *AlbumActor) placeRequestToGetAlbumInfo(r GetAlbumInfoReq) {
	a.inbox.Send(r)
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
