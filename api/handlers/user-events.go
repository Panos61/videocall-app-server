package api

import (
	"net/http"
	"server/internal/userevents"
	"server/internal/utils"
)

func UserEventsHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("room_id")
	if roomID == "" {
		http.Error(w, "room ID is required", http.StatusBadRequest)
		return
	}

	cookie, err := r.Cookie("rsCookie")
	if err != nil {
		http.Error(w, "failed to get jwt cookie", http.StatusUnauthorized)
		return
	}

	claims, err := utils.ValidateToken(cookie.Value)
	if err != nil {
		http.Error(w, "failed to validate jwt cookie", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "failed to upgrade to websocket", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	userevents.UserEventsSubscription(roomID, claims.ParticipantID, conn)
}
