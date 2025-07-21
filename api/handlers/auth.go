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
				http.Error(w, "failed to generate token", http.StatusInternalServerError)
				return
			}
			utils.JSONResponse(w, map[string]string{
				"token": newToken,
			}, http.StatusOK)
			return
		}
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	utils.JSONResponse(w, map[string]string{"token": token}, http.StatusOK)
}

func AuthorizationHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var requestBody struct {
		RoomID string `json:"room_id"`
		Code   string `json:"code"`
	}

	err = json.Unmarshal(body, &requestBody)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	roomID := requestBody.RoomID
	if roomID == "" {
		http.Error(w, "missing room_id", http.StatusBadRequest)
		return
	}
	_, err = room.GetRoom(roomID)
	if err != nil {
		http.Error(w, "room not found", http.StatusNotFound)
		return
	}

	code := requestBody.Code
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}

	matchesFormat := room.MatchesFormat(code)
	if !matchesFormat {
		utils.JSONResponse(w, map[string]bool{
			"isValid": false,
		}, http.StatusOK)
		return
	}

	hasExpired, err := room.HasExpired(roomID, code)
	if err != nil {
		http.Error(w, "failed to check if code has expired", http.StatusInternalServerError)
		return
	}

	if hasExpired {
		utils.JSONResponse(w, map[string]bool{
			"hasExpired": true,
		}, http.StatusOK)
		return
	}

	utils.JSONResponse(w, map[string]bool{
		"isValid":    true,
		"hasExpired": false,
	}, http.StatusOK)
}
