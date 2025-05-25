package ws

import (
	"encoding/json"
	"jammies_streaming/src/db"
	"jammies_streaming/src/models"
	"jammies_streaming/src/types"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Client struct {
	conn   *websocket.Conn
	send   chan []byte
	roomID string
}

var roomMutex = sync.Mutex{}
var clients = make(map[*Client]bool)
var mutex = sync.Mutex{}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func HandleWS(c *gin.Context) {
	roomID := c.Query("room_id")
	if roomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing room_id"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	client := &Client{
		conn:   conn,
		send:   make(chan []byte, 256),
		roomID: roomID,
	}

	roomMutex.Lock()
	if rooms[roomID] == nil {
		rooms[roomID] = make(map[*Client]bool)
	}
	rooms[roomID][client] = true
	roomMutex.Unlock()

	roomStatesMutex.Lock()
	state, ok := roomStates[roomID]
	roomStatesMutex.Unlock()

	if ok {
		jsonMsg, err := json.Marshal(state)
		if err == nil {
			client.send <- jsonMsg
			log.Println("Enviando estado inicial al nuevo cliente:", string(jsonMsg))

		}
	}

	go writePump(client)
	readPump(client)
}

func readPump(client *Client) {
	defer func() {
		removeClient(client)
		client.conn.Close()
	}()

	for {
		var state types.PlayTrackState
		err := client.conn.ReadJSON(&state)
		if err != nil {
			log.Println("Read error:", err)
			break
		}
		if state.Event == "join_room" {
			roomStatesMutex.Lock()
			currentState, ok := roomStates[client.roomID]
			roomStatesMutex.Unlock()

			if ok {
				jsonMsg, err := json.Marshal(currentState)
				if err != nil {
					log.Println("Marshal error:", err)
					continue
				}
				client.send <- jsonMsg
			}
			continue
		}
		if state.Event == "playing" || state.Event == "paused" {
			trackID, err := uuid.Parse(state.TrackID)
			if err != nil {
				log.Println("Invalid UUID:", err)
				continue
			}

			audioURL := GetTrackURL(trackID)

			response := types.PlayTrackGetState{
				Event:       state.Event,
				TrackID:     state.TrackID,
				AudioURL:    audioURL,
				CurrentTime: state.CurrentTime,
			}

			roomStatesMutex.Lock()
			roomStates[client.roomID] = response
			roomStatesMutex.Unlock()

			jsonMsg, err := json.Marshal(response)
			if err != nil {
				log.Println("Marshal error:", err)
				continue
			}

			broadcastToRoom(client.roomID, jsonMsg)
		}
	}
}

func writePump(client *Client) {
	defer client.conn.Close()
	for msg := range client.send {
		err := client.conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			log.Println("Write error:", err)
			break
		}
	}
}

func removeClient(client *Client) {
	roomMutex.Lock()
	defer roomMutex.Unlock()

	if room, ok := rooms[client.roomID]; ok {
		delete(room, client)
		if len(room) == 0 {
			delete(rooms, client.roomID)
		}
	}
}

var (
	trackCache = make(map[string]string)
	cacheMutex = sync.Mutex{}
)

func GetTrackURL(trackID uuid.UUID) string {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	idStr := trackID.String()

	if url, ok := trackCache[idStr]; ok {
		return url
	}

	var track models.Track
	if err := db.DB.First(&track, "id = ?", trackID).Error; err != nil {
		log.Println("Error fetching track:", err)
		return ""
	}

	trackCache[idStr] = track.AudioUrl
	return track.AudioUrl
}
