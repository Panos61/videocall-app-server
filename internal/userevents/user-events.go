package userevents

import (
	"encoding/json"
	"fmt"

	"server/internal/events"
	"server/internal/rdb"

	"github.com/gorilla/websocket"
)

type UserEvent struct {
	Type      string         `json:"type"`
	SessionID string         `json:"session_id,omitempty"`
	Payload   map[string]any `json:"payload"`
}

func UserEventsSubscription(roomID string, participantID string, conn *websocket.Conn) {
	done := make(chan struct{})

	// Start a goroutine to read from the websocket
	// When the connection is closed, this will detect it
	go func() {
		defer close(done)
		for {
			// ReadMessage blocks until a message is received or an error occurs
			var clientEvent UserEvent
			err := conn.ReadJSON(&clientEvent)
			if err != nil {
				return
			}
			// fmt.Println("Reading clientEvent", clientEvent)

			switch clientEvent.Type {
			case events.RaisedHand:
				handleEvent(roomID, participantID, clientEvent.Payload, handleRaisedHandSent)
			case events.ShareScreenStarted:
				handleEvent(roomID, participantID, clientEvent.Payload, handleShareScreenStarted)
			case events.ReactionSent:
				handleEvent(roomID, participantID, clientEvent.Payload, handleReactionSent)
			default:
				fmt.Println("clientEvent.Type", clientEvent.Type)
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
					return
				}
				fmt.Println("Reading msg", msg)

				var eventData UserEvent
				err = json.Unmarshal([]byte(msg.Payload), &eventData)
				if err != nil {
					continue
				}

				fmt.Println("Reading eventData", eventData)

				// Check again if the connection is closed before writing
				select {
				case <-done:
					return
				default:
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

func PublishUserEvent(roomID string, event UserEvent) {
	payloadJSON, err := json.Marshal(event)
	if err != nil {
		return
	}
	fmt.Printf("Publishing user_event: %s\n", string(payloadJSON))

	if err := rdb.Client().Publish(rdb.Context(), "room:"+roomID+":user_events", payloadJSON).Err(); err != nil {
		return
	}
}
