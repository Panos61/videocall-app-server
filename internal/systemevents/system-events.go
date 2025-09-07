package systemevents

import (
	"encoding/json"
	"fmt"

	"server/internal/events"
	"server/internal/rdb"

	"github.com/gorilla/websocket"
)

type SystemEvent struct {
	Type      string         `json:"type"`
	SessionID string         `json:"session_id"`
	Payload   map[string]any `json:"payload"`
}

func SystemEventsSubscription(roomID string, participantID string, conn *websocket.Conn) {
	done := make(chan struct{})

	// Start a goroutine to read from the websocket
	// When the connection is closed, this will detect it
	go func() {
		defer close(done)
		for {
			// ReadMessage blocks until a message is received or an error occurs
			var clientEvent SystemEvent
			err := conn.ReadJSON(&clientEvent)
			if err != nil {
				return
			}

			handleClientEvent(roomID, clientEvent)
		}
	}()

	go func() {
		subscriber := rdb.Client().Subscribe(rdb.Context(), "room:"+roomID+":system_events")
		defer subscriber.Close()

		for {
			// Check if the connection is closed
			select {
			case <-done:
				// Connection is closed, exit the goroutine
				return
			default:
				msg, err := subscriber.ReceiveMessage(rdb.Context())
				if err != nil {
					fmt.Println("error receiving message", err)
					return
				}

				var eventData SystemEvent
				err = json.Unmarshal([]byte(msg.Payload), &eventData)
				if err != nil {
					fmt.Println("error unmarshalling message", err)
					continue
				}

				// Check again if the connection is closed before writing
				select {
				case <-done:
					return
				default:
					if shouldSkipNotification(eventData, participantID) {
						continue
					}

					if err := conn.WriteJSON(eventData); err != nil {
						fmt.Println("error writing message", err)
						return
					}
				}
			}
		}
	}()

	// Read messages from the client until connection closes
	<-done
}

func handleClientEvent(roomID string, event SystemEvent) {
	switch event.Type {
	case events.UserJoined:
		publishSystemEvent(roomID, event)
	}
}

func publishSystemEvent(roomID string, event SystemEvent) {
	payloadJSON, err := json.Marshal(event)
	if err != nil {
		return
	}

	if err := rdb.Client().Publish(rdb.Context(), "room:"+roomID+":system_events", payloadJSON).Err(); err != nil {
		return
	}
}

func shouldSkipNotification(event SystemEvent, participantID string) bool {
	storedParticipantID, err := rdb.Client().Get(rdb.Context(), "session:"+event.SessionID).Result()
	if err != nil {
		return true
	}

	return participantID == storedParticipantID
}
