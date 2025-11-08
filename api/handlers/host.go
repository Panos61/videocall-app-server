package api

import (
	"encoding/json"
	"io"
	"net/http"
	"server/internal/room"
	"server/internal/utils"
	"strings"
)

func AssignHostHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("room_id")
	if roomID == "" {
		http.Error(w, "room ID is required", http.StatusBadRequest)
		return
	}

	tokenStr := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	claims, err := utils.ValidateToken(tokenStr)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	if !claims.IsHost {
		http.Error(w, "only host can assign new host", http.StatusForbidden)
		return
	}

	var requestBody struct {
		NewHostID string `json:"new_host_id"`
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

	err = room.SetHost(roomID, requestBody.NewHostID)
	if err != nil {
		http.Error(w, "failed to set new host", http.StatusInternalServerError)
		return
	}

	utils.JSONResponse(w, map[string]string{
		"new_host_id": requestBody.NewHostID,
	}, http.StatusOK)
}
