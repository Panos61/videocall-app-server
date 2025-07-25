package api

import (
	"net/http"
	"server/internal/call"
	"server/internal/utils"
	"strings"
)

func StartCallHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("room_id")
	if roomID == "" {
		http.Error(w, "room ID is required", http.StatusBadRequest)
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

	if !claims.IsHost {
		http.Error(w, "only host can start call", http.StatusForbidden)
		return
	}

	callState, err := call.StartCall(roomID, claims.ParticipantID)
	if err != nil {
		http.Error(w, "failed to start call", http.StatusInternalServerError)
		return
	}

	utils.JSONResponse(w, callState, http.StatusOK)
}

func LeaveCallHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("room_id")

	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	claims, err := utils.ValidateToken(token)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	_, err = call.LeaveCall(roomID, claims.ParticipantID)
	if err != nil {
		utils.JSONResponse(w, map[string]bool{
			"leftCall": false,
		}, http.StatusBadRequest)
		return
	}

	utils.JSONResponse(w, map[string]bool{
		"leftCall": true,
	}, http.StatusOK)
}

func GetCallStateHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("room_id")
	if roomID == "" {
		http.Error(w, "room ID is required", http.StatusBadRequest)
		return
	}

	callState, err := call.GetCallState(roomID)
	if err != nil {
		http.Error(w, "failed to get call state", http.StatusInternalServerError)
		return
	}

	utils.JSONResponse(w, callState, http.StatusOK)
}

func CallBroadcastHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("room_id")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "failed to upgrade to websocket", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	call.CallSubscription(roomID, conn)
}
