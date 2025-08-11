package api

import (
	"log"
	"net/http"
	"server/internal/events"
	"server/internal/userevents"
	"server/internal/utils"
	"server/internal/websocket"

	gorilla_websocket "github.com/gorilla/websocket"
)

var (
	userEventConnPool = websocket.NewWSConnectionPool()
	eventRegistry     = events.NewEventRegistry()
)

func init() {
	reactionHandler := userevents.NewReactionHandler(userEventConnPool)
	eventRegistry.RegisterHandler(reactionHandler)

	raisedHandHandler := userevents.NewRaisedHandHandler(userEventConnPool)
	eventRegistry.RegisterHandler(raisedHandHandler)
}

func UserEventHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("room_id")
	if roomID == "" {
		http.Error(w, "room id is required", http.StatusBadRequest)
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

	connection := &websocket.WSConnection{Socket: conn}
	userEventConnPool.AddConnection(roomID, claims.ParticipantID, connection)

	defer func() {
		userEventConnPool.RemoveConnection(roomID, claims.ParticipantID)
	}()

	for {
		var baseEvent events.BaseEvent
		if err := conn.ReadJSON(&baseEvent); err != nil {
			if gorilla_websocket.IsCloseError(err, gorilla_websocket.CloseGoingAway, gorilla_websocket.CloseAbnormalClosure) {
				log.Printf("connection closed by client: %v", err)
				return
			}
			log.Printf("error reading event: %v", err)
			continue
		}

		baseEvent.RoomID = roomID
		baseEvent.SenderID = claims.ParticipantID

		if err := eventRegistry.HandleEvent(baseEvent); err != nil {
			log.Printf("Failed to handle event %s from %s: %v", baseEvent.Type, claims.ParticipantID, err)

			errorMsg := map[string]any{
				"type":  "error",
				"error": err.Error(),
			}
			connection.Send(errorMsg)
		}
	}
}
