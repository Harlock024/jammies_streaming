package ws

func broadcastToRoom(roomID string, msg []byte) {
	roomMutex.Lock()
	clients, ok := rooms[roomID]
	if !ok {
		roomMutex.Unlock()
		return
	}

	clientList := make([]*Client, 0, len(clients))
	for client := range clients {
		clientList = append(clientList, client)
	}
	roomMutex.Unlock()

	for _, client := range clientList {
		select {
		case client.send <- msg:
		default:
			roomMutex.Lock()
			close(client.send)
			delete(rooms[roomID], client)
			roomMutex.Unlock()
		}
	}
}
