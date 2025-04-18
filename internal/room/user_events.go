package room

import (
	"encoding/json"
	"fmt"

	"server/internal/participant"
	"server/internal/rdb"

	"github.com/gorilla/websocket"
)

type UserEvent struct {
	Type    string         `json:"type"`
	Payload map[string]any `json:"payload"`
}

func UserEventSubscription(roomID string, participantID string, conn *websocket.Conn) {
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
				fmt.Println("read error:", err)
				return
			}
		}
	}()

	go func() {
		subscriber := rdb.Client().Subscribe(rdb.Context(), "room:"+roomID+":user_events")
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

				fmt.Println("message received", msg.Payload)

				var eventData UserEvent
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

func notifyUserLeft(roomID string, p *participant.Participant) (bool, error) {
	payload := UserEvent{
		Type: "user_left",
		Payload: map[string]any{
			"participant_id":   p.ID,
			"participant_name": p.Username,
		},
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return false, err
	}

	if err := rdb.Client().Publish(rdb.Context(), "room:"+roomID+":user_events", payloadJSON).Err(); err != nil {
		return false, err
	}

	return true, nil
}

func shouldSkipNotification(event UserEvent, participantID string) bool {
	return event.Payload["participant_id"] == participantID
}
