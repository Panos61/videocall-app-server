package systemevents

import (
	"encoding/json"
	"fmt"

	"server/internal/rdb"

	"github.com/gorilla/websocket"
)

type SystemEvent struct {
	Type    string         `json:"type"`
	Payload map[string]any `json:"payload"`
}

func SystemEventsSubscription(roomID string, participantID string, conn *websocket.Conn) {
	done := make(chan struct{})

	// Start a goroutine to read from the websocket
	// When the connection is closed, this will detect it
	go func() {
		defer close(done)
		for {
			// ReadMessage blocks until a message is received or an error occurs
			_, _, err := conn.ReadMessage()
			if err != nil {
				// Connection was closed or there was an error
				return
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

func shouldSkipNotification(event SystemEvent, participantID string) bool {
	return event.Payload["participant_id"] == participantID
}
