package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
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

	hostParticipant, err := room.SetHostParticipant(newRoom.ID)
	if err != nil {
		http.Error(w, "failed to set host participant", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"id":           newRoom.ID,
		"participants": *hostParticipant,
	}

	utils.JSONResponse(w, response, http.StatusCreated)
}

func JoinRoomHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var requestBody struct {
		KeyInput string `json:"invKey"`
	}

	err = json.Unmarshal(body, &requestBody)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	isAuthorized, roomID, err := room.AuthorizeInvitationKey(requestBody.KeyInput)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"isAuthorized": false,
			"roomID":       "",
		})
		return
	}

	_, err = room.GetRoom(roomID)
	if err != nil {
		http.Error(w, "room not found", http.StatusNotFound)
		return
	}

	var participant *room.Participant

	if isAuthorized {
		participant, err = room.JoinRoom(roomID)
		if err != nil {
			http.Error(w, "failed to join room", http.StatusInternalServerError)
		}
	}

	response := struct {
		RoomID      string            `json:"room_id"`
		Participant *room.Participant `json:"participant"`
	}{
		RoomID:      roomID,
		Participant: participant,
	}

	utils.JSONResponse(w, response, http.StatusOK)
}

func LeaveRoomHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("id")

	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	claims, err := utils.ValidateToken(token)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	leftRoom, err := room.LeaveRoom(roomID, claims.ParticipantID)
	log.Printf("err leaving %v", err)
	if err != nil {
		utils.JSONResponse(w, map[string]bool{
			"leftRoom": false,
		}, http.StatusBadRequest)
		return
	}

	fmt.Printf("left: %v\n", leftRoom)

	utils.JSONResponse(w, map[string]bool{
		"leftRoom": true,
	}, http.StatusOK)
}

func GetCallParticipantsHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("room_id")
	if roomID == "" {
		return
	}

	participantData, err := room.GetCallParticipants(roomID)
	if err != nil {
		http.Error(w, "failed to get participants", http.StatusBadRequest)
	}

	utils.JSONResponse(w, map[string]any{"roomParticipants": participantData}, http.StatusOK)
}
