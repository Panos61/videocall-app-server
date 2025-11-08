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
	SessionID string         `json:"session_id,omitempty"`
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

			switch clientEvent.Type {
			case events.HostLeft:
				handleTypedEvent(roomID, clientEvent.Payload, handleHostLeft)
			case events.HostUpdated:
				handleTypedEvent(roomID, clientEvent.Payload, handleHostUpdated)
			default:
				fmt.Println("clientEvent.Type", clientEvent.Type)
			}
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
					return
				}

				var eventData SystemEvent
				err = json.Unmarshal([]byte(msg.Payload), &eventData)
				if err != nil {
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
						return
					}
				}
			}
		}
	}()

	// Read messages from the client until connection closes
	<-done
}

func PublishSystemEvent(roomID string, event SystemEvent) {
	payloadJSON, err := json.Marshal(event)
	if err != nil {
		return
	}
	fmt.Printf("Publishing system_event: %s\n", string(payloadJSON))

	if err := rdb.Client().Publish(rdb.Context(), "room:"+roomID+":system_events", payloadJSON).Err(); err != nil {
		return
	}
}

func shouldSkipNotification(event SystemEvent, participantID string) bool {
	if event.SessionID == "" {
		return false
	}

	storedParticipantID, err := rdb.Client().Get(rdb.Context(), "session:"+event.SessionID).Result()
	if err != nil {
		return true
	}

	return participantID == storedParticipantID
}
