package signalling

import (
	"log"
	"net/http"
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
	Type        string `json:"type"`
	SessionID   string `json:"sessionID"`
	Description string `json:"description,omitempty"`
	To          string `json:"to,omitempty"`
}

var rooms = make(map[string]map[string]*Connection)
var roomsMutex sync.Mutex

func (c *Connection) Send(message Message) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Socket.WriteJSON(message)
}

func SignallingHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
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

	roomsMutex.Lock()
	if rooms[roomID] == nil {
		rooms[roomID] = make(map[string]*Connection)
	}
	participants := rooms[roomID]
	roomsMutex.Unlock()

	var message Message
	for {
		err = conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		// Add the participant to the room if not already present
		if participants[message.SessionID] == nil {
			connection := &Connection{Socket: conn}
			participants[message.SessionID] = connection
		}

		switch message.Type {
		case "connect":
			// Notify all other participants that a new user has joined
			for userID, client := range participants {
				// if userID != message.SessionID { // Skip the sender
				err := client.Send(Message{
					Type:      "session_joined",
					SessionID: message.SessionID,
				})
				if err != nil {
					log.Printf("WebSocket error: %s", err)
					client.Socket.Close()
					delete(participants, userID)
				}
				// }
			}

			err := conn.WriteJSON(Message{
				Type:      "session_joined",
				SessionID: message.SessionID,
			})

			if err != nil {
				log.Printf("WebSocket error: %s", err)
				delete(participants, message.SessionID)
			}

		case "disconnect":
			log.Printf("Participant %s disconnected!", message.SessionID)

			for userID, client := range participants {
				if userID != message.SessionID {
					err := client.Send(Message{
						Type:      "disconnect",
						SessionID: message.SessionID,
					})
					if err != nil {
						client.Socket.Close()
						delete(participants, userID)
					}
				}
			}

			delete(participants, message.SessionID)

		default:
			for userID, client := range participants {
				if userID != message.SessionID {
					err := client.Send(message)
					if err != nil {
						log.Printf("WebSocket error: %s", err)
						client.Socket.Close()
						delete(participants, userID)
					}
				}
			}
		}
	}
}
