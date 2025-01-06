package api

import (
	"fmt"
	"net/http"
	"server/internal/room"
	"server/internal/utils"
	"time"
)

func SetInvitationHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("id")

	existingRoom, err := room.GetRoom(roomID)
	if existingRoom == nil || err != nil {
		http.Error(w, "room not found", http.StatusNotFound)
		return
	}

	invitation := room.GenerateInvitationCode(existingRoom.ID)
	invitationURL, err := room.SetInvitation(existingRoom.ID, invitation)
	if err != nil {
		http.Error(w, "failed to set invKey to this room", http.StatusInternalServerError)
		return
	}

	err = room.CreateInvitationIndex(invitation, existingRoom.ID)
	if err != nil {
		http.Error(w, "failed to create reverse index for invitation key.", http.StatusInternalServerError)
		return
	}

	utils.JSONResponse(w, map[string]string{
		"invitation": invitationURL,
	}, http.StatusOK)
}

// Server-sent event handler to check for expired invitation key
func SSEInvitationHandler(w http.ResponseWriter, r *http.Request) {
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
			return

		default:
			isExpired, err := room.IsInvitationExpired(roomID)
			if err != nil {
				fmt.Fprintf(w, "event: error\ndata: %v\n\n", err)
				flusher.Flush()
				return
			}

			if isExpired {
				newCode := room.GenerateInvitationCode(roomID)

				invitationURL, err := room.SetInvitation(roomID, newCode)
				if err != nil {
					fmt.Fprintf(w, "event: error\ndata: %v\n\n", err)
					flusher.Flush()
					return
				}

				fmt.Fprintf(w, "event: update\ndata: %s\n\n", invitationURL)
				flusher.Flush()

				err = room.CreateInvitationIndex(newCode, roomID)
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
