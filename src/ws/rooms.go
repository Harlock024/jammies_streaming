package ws

import (
	"encoding/json"
	"jammies_streaming/src/types"
	"log"
	"sync"
)

var rooms = make(map[string]map[*Client]bool)
var roomStates = make(map[string]types.PlayTrackGetState)
var roomStatesMutex sync.Mutex

func joinRoom(client *Client, roomID string) {
	mutex.Lock()
	defer mutex.Unlock()

	client.roomID = roomID
	if rooms[roomID] == nil {
		rooms[roomID] = make(map[*Client]bool)
	}
	rooms[roomID][client] = true
}
func sendCurrentStateToClient(client *Client) {
	roomMutex.Lock()
	state, exists := roomStates[client.roomID]
	roomMutex.Unlock()

	if !exists {
		return
	}

	jsonMsg, err := json.Marshal(state)
	if err != nil {
		log.Println("Marshal error:", err)
		return
	}

	select {
	case client.send <- jsonMsg:
	default:
	}
}

func updateRoomStateAndBroadcast(roomID string, state types.PlayTrackGetState) {
	mutex.Lock()
	defer mutex.Unlock()

	roomStates[roomID] = state

	for client := range rooms[roomID] {
		select {
		case client.send <- mustMarshal(state):
		default:
			close(client.send)
			delete(rooms[roomID], client)
		}
	}
}

func mustMarshal(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		log.Println("Marshal error:", err)
		return nil
	}
	return b
}
