package api

import (
	"encoding/json"
	"io"
	"net/http"
	"server/internal/room"
	"server/internal/utils"
	"strings"
)

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

func AuthorizeInvitationHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var requestBody struct {
		KeyInput string `json:"keyInput"`
	}

	err = json.Unmarshal(body, &requestBody)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	isAuthorized, roomID, err := room.AuthorizeInvitationKey(requestBody.KeyInput)
	if err != nil {
		utils.JSONResponse(w, map[string]interface{}{
			"isAuthorized": false,
			"roomID":       "",
		}, http.StatusUnauthorized)
		return
	}

	utils.JSONResponse(w, map[string]interface{}{
		"isAuthorized": isAuthorized,
		"roomID":       roomID,
	}, http.StatusOK)
}
