package api

import (
	"encoding/json"
	"io"
	"net/http"
	"server/internal/participant"
	"server/internal/room"
	"server/internal/utils"
	"strings"
)

type reqBody struct {
	Username  string `json:"username"`
	AvatarSrc string `json:"avatar_src"`
}

type StartCallResponse struct {
	RoomID      string                   `json:"room_id"`
	Participant *participant.Participant `json:"participant"`
}

func StartCallHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("room_id")

	_, err := room.GetRoom(roomID)
	if err != nil {
		http.Error(w, "room not found", http.StatusNotFound)
		return
	}

	var payload reqBody

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(body, &payload)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if token == "" {
		http.Error(w, "authorization token missing", http.StatusBadRequest)
		return
	}

	claims, err := utils.ValidateToken(token)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	participant, err := room.StartCall(roomID, claims.ParticipantID, payload.Username, payload.AvatarSrc)
	if err != nil {
		utils.JSONResponse(w, map[string]string{"error": "failed to start call"}, http.StatusBadRequest)
		return
	}

	response := StartCallResponse{
		RoomID:      roomID,
		Participant: participant,
	}

	utils.JSONResponse(w, response, http.StatusOK)
}
