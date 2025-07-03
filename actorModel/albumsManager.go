package actorModel

import (
	"time"

	"example.com/web-service-gin/monitoredChannel"
	"gorm.io/gorm"
)

const MAX_ENTRIES_FOR_ALBUM_MANAGER_INBOX = 1_000

type AlbumManager struct {
	albumActors map[AlbumId]*AlbumActor
	inbox       *monitoredChannel.MonitoredChannel[RequestToGetAlbumInfo]
}

type RequestToGetAlbumInfo struct {
	AlbumIdToRequest AlbumId
	UsingDb          *gorm.DB
	PlaceInfoHere    chan *ResponseAfterGettingAlbumInfo
}

func CreateNewAlbumManager() *AlbumManager {
	newAlbumManager := &AlbumManager{
		albumActors: make(map[AlbumId]*AlbumActor),
		inbox:       monitoredChannel.NewMonitoredChannel[RequestToGetAlbumInfo]("album_manager_inbox", MAX_ENTRIES_FOR_ALBUM_MANAGER_INBOX, 1*time.Second),
	}

	go newAlbumManager.startProcessingMessagesFromInbox()

	return newAlbumManager
}

func (m *AlbumManager) PlaceRequestToGetAlbumInfo(r RequestToGetAlbumInfo) {
	m.inbox.Send(r)
}

func (m *AlbumManager) startProcessingMessagesFromInbox() {
	for {
		m.processRequestToGetAlbumInfo(m.inbox.Receive())
	}
}

func (m *AlbumManager) processRequestToGetAlbumInfo(r RequestToGetAlbumInfo) {
	_, albumActorWasFound := m.albumActors[r.AlbumIdToRequest]
	if !albumActorWasFound {
		maybeActorCreated, err := createAlbumActor(r.AlbumIdToRequest, r.UsingDb)
		if err != nil {
			r.PlaceInfoHere <- &ResponseAfterGettingAlbumInfo{MaybeAlbumInfoFound: nil, MaybeErrorFound: err}
			return
		}

		m.albumActors[r.AlbumIdToRequest] = maybeActorCreated
	}

	m.albumActors[r.AlbumIdToRequest].placeRequestToGetAlbumInfo(GetAlbumInfoReq{placeInfoHere: r.PlaceInfoHere})
}
