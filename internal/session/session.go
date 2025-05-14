package session

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgraderV2 = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type MessageV2 struct {
	Type        string `json:"type"`
	RoomID      string `json:"roomID"`
	SessionID   string `json:"sessionID"`
	Token       string `json:"token"`
	Description string `json:"description"`
	To          string `json:"to"`
}

type ConnectionV2 struct {
	Socket *websocket.Conn
	mu     sync.Mutex
}

var rooms = make(map[string]map[string]*ConnectionV2)
var roomsMutex sync.Mutex

func (c *ConnectionV2) Send(message MessageV2) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Socket.WriteJSON(message)
}

func SessionHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgraderV2.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal("Error handling websocket connection.")
		return
	}

	defer conn.Close()

	roomID := r.PathValue("room_id")
	if roomID == "" {
		log.Println("Room ID is missing.")
		return
	}

	var connected bool
	var sessionID string

	for {
		var msg MessageV2
		if err := conn.ReadJSON(&msg); err != nil {
			break
		}

		if !connected {
			sessionID = msg.SessionID
			connected = true

			roomsMutex.Lock()
			if rooms[roomID] == nil {
				rooms[roomID] = make(map[string]*ConnectionV2)
			}
			rooms[roomID][sessionID] = &ConnectionV2{Socket: conn}
			roomsMutex.Unlock()

			livekitToken, err := createLivekitToken(roomID, sessionID)
			if err != nil {
				break
			}

			fmt.Println("livekitToken", livekitToken)

			conn.WriteJSON(MessageV2{
				Type:        "livekit_token",
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
