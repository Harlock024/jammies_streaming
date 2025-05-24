package ws

func StartBroadcast() {
	for {
		msg := <-broadcast
		mutex.Lock()
		for client := range clients {
			select {
			case client.send <- msg:
			default:
				close(client.send)
				delete(clients, client)
			}
		}
		mutex.Unlock()
	}
}
