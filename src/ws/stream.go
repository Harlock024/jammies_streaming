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
	conn *websocket.Conn
	send chan []byte
}

var clients = make(map[*Client]bool)
var broadcast = make(chan []byte)
var mutex = sync.Mutex{}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func HandleWS(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	client := &Client{
		conn: conn,
		send: make(chan []byte, 256),
	}

	mutex.Lock()

	clients[client] = true

	mutex.Unlock()

	go writePump(client)
	readPump(client)

}

func readPump(client *Client) {
	defer func() {
		mutex.Lock()
		delete(clients, client)
		mutex.Unlock()
		client.conn.Close()
	}()

	for {
		var state types.PlayTrackState
		err := client.conn.ReadJSON(&state)
		if err != nil {
			log.Println("Read error:", err)
			break
		}

		audioURL := GetTrackURL(state.TrackID)

		response := types.PlayTrackGetState{
			Event:       state.Event,
			TrackID:     state.TrackID,
			AudioURL:    audioURL,
			CurrentTime: state.CurrentTime,
		}

		jsonMsg, err := json.Marshal(response)
		if err != nil {
			log.Println("Marshal error:", err)
			continue
		}
		broadcast <- jsonMsg
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

func GetTrackURL(trackID uuid.UUID) string {

	var track models.Track
	if err := db.DB.First(&track, trackID).Error; err != nil {
		log.Println("Error fetching track:", err)
		return ""
	}

	return track.AudioUrl
}
