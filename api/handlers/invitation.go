package api

import (
	"fmt"
	"net/http"
	"server/internal/room"
	"server/internal/utils"
	"time"
)

func SetInvKeyHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("id")

	existingRoom, err := room.GetRoom(roomID)
	if existingRoom == nil || err != nil {
		http.Error(w, "room not found", http.StatusNotFound)
		return
	}

	invKey := room.GenerateInvKey(existingRoom.ID)
	err = room.SetRoomKey(existingRoom.ID, invKey)
	if err != nil {
		http.Error(w, "failed to set invKey to this room", http.StatusInternalServerError)
		return
	}

	err = room.InvitationKeyReverseIndex(invKey, existingRoom.ID)
	if err != nil {
		http.Error(w, "failed to create reverse index for invitation key.", http.StatusInternalServerError)
		return
	}

	utils.JSONResponse(w, map[string]string{
		"invitation_key": invKey,
	}, http.StatusOK)
}

// Server-sent event handler to check for expired invitation key
func SSEKeyUpdateHandler(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	roomID := r.PathValue("id")

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Listen for client disconnect
	notify := r.Context().Done()

	for {
		select {
		case <-notify:
			// log.Printf("Client disconnected from room: %s", roomID)
			return

		default:
			isExpired, err := room.IsKeyExpired(roomID)
			if err != nil {
				fmt.Fprintf(w, "event: error\ndata: %v\n\n", err)
				flusher.Flush()
				return
			}

			if isExpired {
				newKey := room.GenerateInvKey(roomID)
				err := room.SetRoomKey(roomID, newKey)
				if err != nil {
					fmt.Fprintf(w, "event: error\ndata: %v\n\n", err)
					flusher.Flush()
					return
				}

				fmt.Fprintf(w, "event: update\ndata: %s\n\n", newKey)
				flusher.Flush()

				err = room.InvitationKeyReverseIndex(newKey, roomID)
				if err != nil {
					fmt.Fprintf(w, "event: error\ndata: %v\n\n", err)
					flusher.Flush()
					return
				}
			}

			time.Sleep(5 * time.Second)
		}
	}
}
