package api

import (
	"encoding/json"
	"io"
	"net/http"
	"server/internal/participant"
	"server/internal/room"
	"server/internal/utils"
)

func AssignHostHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("room_id")
	if roomID == "" {
		http.Error(w, "room ID is required", http.StatusBadRequest)
		return
	}

	var requestBody struct {
		CurrentHostID string `json:"current_host_id"`
		NewHostID     string `json:"new_host_id"`
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if err := json.Unmarshal(body, &requestBody); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	currentHost, err := participant.GetParticipant(roomID, requestBody.CurrentHostID)
	if err != nil {
		http.Error(w, "failed to get current host", http.StatusInternalServerError)
		return
	}

	if !currentHost.IsHost {
		http.Error(w, "only host can assign new host", http.StatusForbidden)
		return
	}

	err = room.SetHost(roomID, requestBody.CurrentHostID, requestBody.NewHostID)
	if err != nil {
		http.Error(w, "failed to set new host", http.StatusInternalServerError)
		return
	}

	utils.JSONResponse(w, map[string]string{
		"new_host_id": requestBody.NewHostID,
	}, http.StatusOK)
}
