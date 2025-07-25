package api

import (
	"net/http"
	"server/internal/participant"
	"server/internal/room"
	"server/internal/utils"
	"strings"
)

func CreateRoomHandler(w http.ResponseWriter, r *http.Request) {
	newRoom, err := room.CreateRoom()
	if err != nil {
		http.Error(w, "failed to create room", http.StatusInternalServerError)
		return
	}

	hostParticipant, err := room.SetHostParticipant(newRoom)
	if err != nil {
		http.Error(w, "failed to set host participant", http.StatusInternalServerError)
		return
	}

	// set host's jwt cookie
	jwtCookie := http.Cookie{
		Name:     "jwt-cookie",
		Value:    hostParticipant.Token,
		MaxAge:   3600,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		Domain:   "",
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &jwtCookie)

	response := map[string]interface{}{
		"id":           newRoom,
		"participants": *hostParticipant,
	}

	utils.JSONResponse(w, response, http.StatusCreated)
}

func JoinRoomHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("room_id")
	if roomID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err := room.GetRoom(roomID)
	if err != nil {
		http.Error(w, "room not found", http.StatusNotFound)
		return
	}

	var joinedParticipant *participant.Participant

	joinedParticipant, err = room.JoinRoom(roomID)
	if err != nil {
		http.Error(w, "failed to join room", http.StatusInternalServerError)
	}

	jwtCookie := http.Cookie{
		Name:     "jwt-cookie",
		Value:    joinedParticipant.Token,
		MaxAge:   3600,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		Domain:   "",
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &jwtCookie)

	response := struct {
		RoomID      string                   `json:"room_id"`
		Participant *participant.Participant `json:"participant"`
	}{
		RoomID:      roomID,
		Participant: joinedParticipant,
	}

	utils.JSONResponse(w, response, http.StatusOK)
}

// used when user leaves the room by navigating away from the page
// if there's only one participant, delete the room and relevant user data
// if there's more than one participant, delete the participant data
func ExitRoomHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("room_id")
	if roomID == "" {
		http.Error(w, "room ID is required", http.StatusBadRequest)
		return
	}

	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	claims, err := utils.ValidateToken(token)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	isDeleted, err := room.ExitRoom(roomID, claims.ParticipantID, claims.IsHost)
	if err != nil {
		http.Error(w, "failed to delete room and relevant user data", http.StatusInternalServerError)
		return
	}

	utils.JSONResponse(w, map[string]bool{
		"deleted": isDeleted,
	}, http.StatusOK)
}

// for now it's just the room created_at timestamp
func GetRoomInfoHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("room_id")
	if roomID == "" {
		http.Error(w, "room ID is required", http.StatusBadRequest)
		return
	}

	roomInfo, err := room.GetInfo(roomID)
	if err != nil {
		http.Error(w, "failed to get room info", http.StatusInternalServerError)
		return
	}

	utils.JSONResponse(w, roomInfo.CreatedAt, http.StatusOK)
}

func GetCallParticipantsHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("room_id")
	if roomID == "" {
		return
	}

	participantData, err := participant.GetCallParticipants(roomID)
	if err != nil {
		http.Error(w, "failed to get participants", http.StatusBadRequest)
	}

	utils.JSONResponse(w, map[string]any{"roomParticipants": participantData}, http.StatusOK)
}
