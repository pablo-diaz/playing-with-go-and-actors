package actorModel

import (
	"gorm.io/gorm"
)

const MAX_ENTRIES_FOR_ALBUM_MANAGER_INBOX = 100_000

type AlbumManager struct {
	albumActors map[AlbumId]*AlbumActor
	inbox       chan any
}

type RequestToGetAlbumInfo struct {
	AlbumIdToRequest AlbumId
	UsingDb          *gorm.DB
	PlaceInfoHere    chan *ResponseAfterGettingAlbumInfo
}

func CreateNewAlbumManager() *AlbumManager {
	newAlbumManager := &AlbumManager{
		albumActors: make(map[AlbumId]*AlbumActor),
		inbox:       make(chan any, MAX_ENTRIES_FOR_ALBUM_MANAGER_INBOX),
	}

	go newAlbumManager.startProcessingMessagesFromInbox()

	return newAlbumManager
}

func (m *AlbumManager) PlaceRequestToGetAlbumInfo(r RequestToGetAlbumInfo) {
	m.inbox <- r
}

func (m *AlbumManager) startProcessingMessagesFromInbox() {
	for message := range m.inbox {
		switch request := message.(type) {
		case RequestToGetAlbumInfo:
			m.processRequestToGetAlbumInfo(request)
		case AlbumId:
			m.removeAlbumActor(request)
		}
	}
}

func (m *AlbumManager) removeAlbumActor(idToRemove AlbumId) {
	delete(m.albumActors, idToRemove)
}

func (m *AlbumManager) processRequestToGetAlbumInfo(r RequestToGetAlbumInfo) {
	_, albumActorWasFound := m.albumActors[r.AlbumIdToRequest]
	if !albumActorWasFound {
		maybeActorCreated, err := createAlbumActor(r.AlbumIdToRequest, r.UsingDb, m.inbox)
		if err != nil {
			r.PlaceInfoHere <- &ResponseAfterGettingAlbumInfo{MaybeAlbumInfoFound: nil, MaybeErrorFound: err}
			return
		}

		m.albumActors[r.AlbumIdToRequest] = maybeActorCreated
	}

	m.albumActors[r.AlbumIdToRequest].placeRequestToGetAlbumInfo(GetAlbumInfoReq{placeInfoHere: r.PlaceInfoHere})
}
