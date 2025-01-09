package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"server/internal/room"
	"server/internal/utils"
	"time"
)

func SetInvitationHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("room_id")

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

	roomID := r.PathValue("room_id")

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

			time.Sleep(10 * time.Second)
		}
	}
}

func InvitationSettingsHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("room_id")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var reqBody struct {
		ExpiresIn string `json:"invitation_expiry"`
	}

	err = json.Unmarshal(body, &reqBody)
	fmt.Println(err)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	expirationSet, err := room.SetInvitationExpiration(roomID, reqBody.ExpiresIn)
	fmt.Printf("expirationSet %t\n", expirationSet)

	if err != nil {
		fmt.Println(err)
		utils.JSONResponse(w, map[string]interface{}{
			"expirationSet": false,
			"error":         err.Error(),
		}, http.StatusInternalServerError)
		return
	}

	utils.JSONResponse(w, map[string]interface{}{
		"expirationSet": expirationSet,
	}, http.StatusOK)
}
