package settings

import (
	"encoding/json"
	"fmt"
	"server/internal/rdb"
	"strconv"

	"github.com/gorilla/websocket"
)

type Settings struct {
	InvitationExpiry string `json:"invitation_expiry"`
	InvitePermission bool   `json:"invite_permission"`
}

func GetRoomSettings(roomID string) (Settings, error) {
	settingsData, err := rdb.Client().HGetAll(rdb.Context(), "room:"+roomID+":settings").Result()
	if err != nil {
		return Settings{}, err
	}

	invitationExpiry := settingsData["invitation_expiry"]
	invitePermission, err := strconv.ParseBool(settingsData["invite_permission"])

	if err != nil {
		return Settings{
			InvitationExpiry: invitationExpiry,
			InvitePermission: false,
		}, err
	}

	settings := Settings{
		InvitationExpiry: invitationExpiry,
		InvitePermission: invitePermission,
	}

	return settings, nil
}

func UpdateRoomSettings(roomID string, settings Settings) (Settings, error) {
	_, err := rdb.Client().HSet(rdb.Context(), "room:"+roomID+":settings", map[string]any{
		"invitation_expiry": settings.InvitationExpiry,
		"invite_permission": settings.InvitePermission,
	}).Result()

	if err != nil {
		return Settings{}, fmt.Errorf("error updating room settings: %w", err)
	}

	broadcastSettingsUpdate(roomID, settings)
	return settings, nil
}

// Subscribes to the host-only settings broadcast channel
func SettingsSubscription(roomID string, conn *websocket.Conn) {
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
		subscriber := rdb.Client().Subscribe(rdb.Context(), "room:"+roomID+":settings_broadcast")
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

				var settingsData Settings
				err = json.Unmarshal([]byte(msg.Payload), &settingsData)
				if err != nil {
					fmt.Println("error unmarshalling message", err)
					continue
				}

				// Check again if the connection is closed before writing
				select {
				case <-done:
					return
				default:
					if err := conn.WriteJSON(settingsData); err != nil {
						fmt.Println("error writing message", err)
						return
					}
				}
			}
		}
	}()

	<-done
}

// Broadcasts the host-only settings update to all connected clients in the room
func broadcastSettingsUpdate(roomID string, settings Settings) (bool, error) {
	payload := Settings{
		InvitationExpiry: settings.InvitationExpiry,
		InvitePermission: settings.InvitePermission,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return false, err
	}

	if err := rdb.Client().Publish(rdb.Context(), "room:"+roomID+":settings_broadcast", payloadJSON).Err(); err != nil {
		return false, err
	}

	return true, nil
}
