package types

import (
	"sync"

	"github.com/google/uuid"
)

type RoomState struct {
	TrackID     uuid.UUID
	Event       string
	CurrentTime float64
	AudioURL    string
}

var (
	RoomsState      = make(map[string]*RoomState)
	RoomsStateMutex sync.Mutex
)
