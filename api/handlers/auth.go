package api

import (
	"encoding/json"
	"io"
	"net/http"
	"server/internal/room"
	"server/internal/utils"
)

func ValidateInvitationHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var requestBody struct {
		RoomID string `json:"room_id"`
		Code   string `json:"code"`
	}

	err = json.Unmarshal(body, &requestBody)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	isExpired, err := room.IsInvitationExpired(requestBody.RoomID)
	if err != nil {
		utils.JSONResponse(w, map[string]interface{}{
			"isValid":   false,
			"isExpired": true,
			"roomID":    "",
		}, http.StatusUnauthorized)
		return
	}

	isValid, roomID, err := room.ValidateInvitation(requestBody.Code)
	if err != nil {
		utils.JSONResponse(w, map[string]interface{}{
			"isValid":   false,
			"isExpired": false,
			"roomID":    "",
		}, http.StatusUnauthorized)
		return
	}

	if roomID != requestBody.RoomID {
		utils.JSONResponse(w, map[string]interface{}{
			"isValid":   false,
			"isExpired": false,
			"roomID":    "",
		}, http.StatusUnauthorized)
		return
	}

	utils.JSONResponse(w, map[string]interface{}{
		"isValid":   isValid,
		"IsExpired": isExpired,
		"roomID":    roomID,
	}, http.StatusOK)
}
