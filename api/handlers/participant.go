package api

import (
	"net/http"
	"server/internal/participant"
	"server/internal/room"
	"server/internal/utils"
	"strings"
)

func GetMeHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("room_id")
	if roomID == "" {
		utils.JSONResponse(w, map[string]interface{}{
			"error": "room not found",
		}, http.StatusNotFound)
		return
	}

	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	claims, err := utils.ValidateToken(token)
	if err != nil {
		utils.JSONResponse(w, map[string]interface{}{
			"error": err.Error(),
		}, http.StatusUnauthorized)
		return
	}

	me, err := participant.GetMe(roomID, claims.ParticipantID)
	if err != nil {
		utils.JSONResponse(w, map[string]interface{}{
			"error": err.Error(),
		}, http.StatusNotFound)
		return
	}

	utils.JSONResponse(w, me, http.StatusOK)
}

func SetSessionHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("room_id")
	if roomID == "" {
		utils.JSONResponse(w, map[string]interface{}{
			"error": "room not found",
		}, http.StatusNotFound)
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
