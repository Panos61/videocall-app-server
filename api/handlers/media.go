package api

import (
	"fmt"
	"log"
	"net/http"
	"server/internal/media"
	"server/internal/utils"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Connection struct {
	Socket *websocket.Conn
	mu     sync.Mutex
}

type Message struct {
	SessionID string           `json:"sessionID"`
	Media     media.MediaState `json:"media"`
}

type AuthMessage struct {
	Type      string `json:"type"`
	Token     string `json:"token"`
	SessionID string `json:"sessionID"`
}

var connectionPool = make(map[string]map[string]*Connection)
var connectionsMutex sync.Mutex

func (c *Connection) Send(message Message) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.Socket.WriteJSON(message)
}

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
	fmt.Println("auth", authMsg)
	if err != nil {
		log.Printf("Error reading auth message: %v", err)
		return
	}

	// Validate JWT
	claims, err := utils.ValidateToken(authMsg.Token)
	if err != nil {
		conn.WriteJSON(map[string]string{"error": "unauthorized"})
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
		fmt.Println("message", message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		updatedMedia, err := media.UpdateMedia(roomID, claims.ParticipantID, message.Media)
		if err != nil {
			log.Printf("Failed to update media state: %v", err)
			continue
		}

		connectionsMutex.Lock()
		for sessionID, client := range participants {
			if sessionID != message.SessionID {
				err := client.Send(Message{
					SessionID: message.SessionID,
					Media:     *updatedMedia,
				})
				if err != nil {
					log.Printf("WebSocket error: %v", err)
					client.Socket.Close()
					delete(participants, sessionID)
				}
			}
		}
		connectionsMutex.Unlock()
	}
}
