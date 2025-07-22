package api

import (
	"log"
	"net/http"

	"server/internal/livekit"
)

func SessionHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal("Error handling websocket connection.")
		return
	}

	defer conn.Close()

	roomID := r.PathValue("room_id")
	if roomID == "" {
		return
	}

	var connected bool
	var sessionID string
	var authMsg AuthMessage

	// Cleanup on disconnect
	defer func() {
		connectionsMutex.Lock()
		delete(connectionPool[roomID], authMsg.SessionID)
		if len(connectionPool[roomID]) == 0 {
			delete(connectionPool, roomID)
		}
		connectionsMutex.Unlock()
	}()

	for {
		var msg Message
		if err := conn.ReadJSON(&msg); err != nil {
			break
		}

		if !connected {
			sessionID = msg.SessionID
			connected = true

			connectionsMutex.Lock()
			if connectionPool[roomID] == nil {
				connectionPool[roomID] = make(map[string]*Connection)
			}
			connectionPool[roomID][sessionID] = &Connection{Socket: conn}
			connectionsMutex.Unlock()

			livekitToken, err := livekit.CreateLivekitToken(roomID, sessionID)
			if err != nil {
				break
			}

			conn.WriteJSON(Message{
				Type:        "livekit_session_token",
				SessionID:   sessionID,
				Token:       livekitToken,
				Description: "User connected to room",
			})

			continue
		}

		if msg.To == "" {
			connectionsMutex.Lock()
			target := connectionPool[roomID][msg.SessionID]
			connectionsMutex.Unlock()

			if target != nil {
				target.Send(msg)
			}
		}
	}

	connectionsMutex.Lock()
	delete(connectionPool[roomID], sessionID)
	connectionsMutex.Unlock()
}
