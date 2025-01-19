package api

import (
	"encoding/json"
	"io"
	"net/http"
	"server/internal/room"
	"server/internal/utils"
	"strings"
)

func RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if token == "" {
		http.Error(w, "token not found", http.StatusUnauthorized)
		return
	}

	claims, err := utils.ValidateToken(token)
	if err != nil {
		if err.Error() == "token is expired" {
			newToken, err := utils.GenerateJWT(claims.ParticipantID, claims.IsHost)
			if err != nil {
				utils.JSONResponse(w, map[string]interface{}{
					"error": "failed to generate token",
				}, http.StatusInternalServerError)
				return
			}
			utils.JSONResponse(w, map[string]interface{}{
				"token": newToken,
			}, http.StatusOK)
			return
		}
		utils.JSONResponse(w, map[string]interface{}{
			"error": "invalid token",
		}, http.StatusUnauthorized)
		return
	}

	utils.JSONResponse(w, map[string]interface{}{
		"token": token,
	}, http.StatusOK)
}

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

	isValid, roomID, err := room.ValidateInvitation(requestBody.Code)
	if err != nil {
		utils.JSONResponse(w, map[string]interface{}{
			"isValid":   false,
			"isExpired": false,
			"roomID":    "",
		}, http.StatusUnauthorized)
		return
	}

	isExpired, err := room.IsExpired(requestBody.RoomID)
	if err != nil {
		utils.JSONResponse(w, map[string]interface{}{
			"isValid":   false,
			"isExpired": true,
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
