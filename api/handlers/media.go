package api

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

func MediaHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("room_id")
	if roomID == "" {
		http.Error(w, "room not found", http.StatusNotFound)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("error handling websocket connection: %v", err)
		return
	}
	defer conn.Close()

	// Wait for auth message
	var authMsg AuthMessage
	err = conn.ReadJSON(&authMsg)
	if err != nil {
		log.Printf("Error reading auth message: %v", err)
		return
	}

	connectionsMutex.Lock()
	if connectionPool[roomID] == nil {
		connectionPool[roomID] = make(map[string]*Connection)
	}
	participants := connectionPool[roomID]
	participants[authMsg.SessionID] = &Connection{Socket: conn}
	connectionsMutex.Unlock()

	// Cleanup on disconnect
	defer func() {
		connectionsMutex.Lock()
		delete(connectionPool[roomID], authMsg.SessionID)
		if len(connectionPool[roomID]) == 0 {
			delete(connectionPool, roomID)
		}
		connectionsMutex.Unlock()
	}()

	var message Message
	for {
		err = conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		connectionsMutex.Lock()
		for sessionID, client := range participants {
			if sessionID != message.SessionID {
				err := client.Send(Message{
					SessionID: message.SessionID,
					Media:     message.Media,
				})
				if err != nil {
					client.Socket.Close()
					delete(participants, sessionID)
				}
			}
		}
		connectionsMutex.Unlock()
	}
}
