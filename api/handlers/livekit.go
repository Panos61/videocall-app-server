package api

import (
	"fmt"
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
		delete(rooms[roomID], authMsg.SessionID)
		if len(rooms[roomID]) == 0 {
			delete(rooms, roomID)
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

			roomsMutex.Lock()
			if rooms[roomID] == nil {
				rooms[roomID] = make(map[string]*Connection)
			}
			rooms[roomID][sessionID] = &Connection{Socket: conn}
			roomsMutex.Unlock()

			livekitToken, err := livekit.CreateLivekitToken(roomID, sessionID)
			if err != nil {
				break
			}

			fmt.Println("livekitToken", livekitToken)

			conn.WriteJSON(Message{
				Type:        "livekit_session_token",
				SessionID:   sessionID,
				Token:       livekitToken,
				Description: "User connected to room",
			})

			continue
		}

		if msg.To == "" {
			roomsMutex.Lock()
			target := rooms[roomID][msg.SessionID]
			roomsMutex.Unlock()

			if target != nil {
				target.Send(msg)
			}
		}
	}

	roomsMutex.Lock()
	delete(rooms[roomID], sessionID)
	roomsMutex.Unlock()
}
