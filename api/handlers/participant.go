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

func GetMeHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("room_id")
	if roomID == "" {
		http.Error(w, "room not found", http.StatusNotFound)
		return
	}

	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	claims, err := utils.ValidateToken(token)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	me, err := participant.GetMe(roomID, claims.ParticipantID)
	if err != nil {
		http.Error(w, "participant not found", http.StatusNotFound)
		return
	}

	utils.JSONResponse(w, me, http.StatusOK)
}

func GetParticipantsHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("room_id")
	if roomID == "" {
		http.Error(w, "room id not found", http.StatusBadRequest)
		return
	}

	allParticipants, participantsInCall, err := participant.GetParticipants(roomID)
	if err != nil {
		http.Error(w, "failed to get participants", http.StatusBadRequest)
		return
	}

	utils.JSONResponse(w, map[string]any{
		"participants":       allParticipants,
		"participantsInCall": participantsInCall,
	}, http.StatusOK)
}

func SetSessionHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("room_id")
	if roomID == "" {
		http.Error(w, "room not found", http.StatusNotFound)
		return
	}

	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")

	claims, err := utils.ValidateToken(token)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	sessionID, err := utils.GenerateSessionID()
	if err != nil {
		http.Error(w, "failed to generate session id", http.StatusBadRequest)
		return
	}

	err = room.StoreUserSession(roomID, sessionID, claims.ParticipantID)
	if err != nil {
		http.Error(w, "failed to store user session id", http.StatusBadRequest)
		return
	}

	utils.JSONResponse(w, sessionID, http.StatusOK)
}

func ParticipantsHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("room_id")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "failed to upgrade to websocket", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	participant.ParticipantsSubscription(roomID, conn)
}

type ParticipantCallData struct {
	RoomID      string                   `json:"room_id"`
	Participant *participant.Participant `json:"participant"`
}

type reqBody struct {
	Username  string `json:"username"`
	AvatarSrc string `json:"avatar_src"`
}

func SetParticipantCallDataHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("room_id")

	_, err := room.GetRoomID(roomID)
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

	participant, err := participant.SetParticipantCallData(roomID, claims.ParticipantID, payload.Username, payload.AvatarSrc)
	if err != nil {
		http.Error(w, "failed to set participant call data", http.StatusBadRequest)
		return
	}

	response := ParticipantCallData{
		RoomID:      roomID,
		Participant: participant,
	}

	utils.JSONResponse(w, response, http.StatusOK)
}
