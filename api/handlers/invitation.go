package api

import (
	"net/http"
	"server/internal/room"
	"server/internal/utils"
)

func GetInvitationHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("room_id")
	if roomID == "" {
		http.Error(w, "missing room_id", http.StatusBadRequest)
		return
	}

	existingRoomID, err := room.GetRoom(roomID)
	if err != nil {
		http.Error(w, "room not found", http.StatusNotFound)
		return
	}

	code, err := room.GetInvitationCode(existingRoomID)
	if err != nil {
		http.Error(w, "failed to set invKey to this room", http.StatusInternalServerError)
		return
	}

	utils.JSONResponse(w, map[string]string{
		"invitation": code,
	}, http.StatusOK)
}
