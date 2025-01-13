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

func UpdateUserMediaHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("room_id")
	_, err := room.GetRoom(roomID)
	if err != nil {
		http.Error(w, "room not found", http.StatusNotFound)
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var reqBody struct {
		Media participant.MediaState `json:"media"`
	}

	err = json.Unmarshal(body, &reqBody)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	claims, err := utils.ValidateToken(token)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	media, err := participant.UpdateUserMediaState(roomID, claims.ParticipantID, participant.MediaState{Audio: reqBody.Media.Audio, Video: reqBody.Media.Video})
	if err != nil {
		http.Error(w, "failed to update media", http.StatusBadRequest)
		return
	}

	utils.JSONResponse(w, media, http.StatusOK)
}
