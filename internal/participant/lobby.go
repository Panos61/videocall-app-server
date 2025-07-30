package participant

import (
	"encoding/json"
	"fmt"
	"server/internal/rdb"

	"github.com/gorilla/websocket"
)

// Subscribes to the participants-in-lobby broadcast channel
func ParticipantsSubscription(roomID string, conn *websocket.Conn) {
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
		subscriber := rdb.Client().Subscribe(rdb.Context(), "room:"+roomID+":participants")
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

				var participantsData []*Participant
				err = json.Unmarshal([]byte(msg.Payload), &participantsData)
				if err != nil {
					fmt.Println("error unmarshalling message", err)
					continue
				}

				// Check again if the connection is closed before writing
				select {
				case <-done:
					return
				default:
					if err := conn.WriteJSON(participantsData); err != nil {
						fmt.Println("error writing message", err)
						return
					}
				}
			}
		}
	}()

	<-done
}

// Broadcasts the participants-in-lobby update to all connected clients in the room
func BroadcastParticipantsUpdate(roomID string, participants []*Participant) (bool, error) {
	payloadJSON, err := json.Marshal(participants)
	if err != nil {
		return false, err
	}

	if err := rdb.Client().Publish(rdb.Context(), "room:"+roomID+":participants", payloadJSON).Err(); err != nil {
		return false, err
	}

	return true, nil
}
