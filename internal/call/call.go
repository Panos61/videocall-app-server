package call

import (
	"encoding/json"
	"fmt"
	"server/internal/rdb"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

type CallState struct {
	RoomID    string    `json:"room_id"`
	IsActive  bool      `json:"is_active"`
	StartedBy string    `json:"started_by"`
	StartedAt time.Time `json:"started_at"`
}

func StartCall(roomID, participantID string) (CallState, error) {
	callState := CallState{
		RoomID:    roomID,
		IsActive:  true,
		StartedBy: participantID,
		StartedAt: time.Now(),
	}

	callStateJSON, err := json.Marshal(callState)
	if err != nil {
		return CallState{}, err
	}

	err = rdb.Client().HSet(rdb.Context(), "call:"+roomID, map[string]any{
		"room_id":    roomID,
		"is_active":  true,
		"started_by": participantID,
		"started_at": time.Now().Unix(),
	}).Err()
	if err != nil {
		return CallState{}, err
	}

	if err = rdb.Client().Publish(rdb.Context(), "room:"+roomID+":call", callStateJSON).Err(); err != nil {
		return CallState{}, err
	}

	return callState, nil
}

func GetCallState(roomID string) (CallState, error) {
	callData, err := rdb.Client().HGetAll(rdb.Context(), "call:"+roomID).Result()
	if err != nil {
		return CallState{}, err
	}

	if len(callData) == 0 {
		return CallState{
			RoomID:    roomID,
			IsActive:  false,
			StartedBy: "",
			StartedAt: time.Time{},
		}, nil
	}

	isActive, _ := strconv.ParseBool(callData["is_active"])
	startedAt, _ := time.Parse(time.RFC3339, callData["started_at"])

	return CallState{
		RoomID:    callData["room_id"],
		IsActive:  isActive,
		StartedBy: callData["started_by"],
		StartedAt: startedAt,
	}, nil
}

// Subscribes to room's call state channel
func CallSubscription(roomID string, conn *websocket.Conn) {
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
		subscriber := rdb.Client().Subscribe(rdb.Context(), "room:"+roomID+":call")
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

				var callState CallState
				err = json.Unmarshal([]byte(msg.Payload), &callState)
				fmt.Println("callState -->", callState)
				if err != nil {
					fmt.Println("error unmarshalling message", err)
					continue
				}

				// Check again if the connection is closed before writing
				select {
				case <-done:
					return
				default:
					if err := conn.WriteJSON(callState); err != nil {
						fmt.Println("error writing message", err)
						return
					}
				}
			}
		}
	}()

	<-done
}
