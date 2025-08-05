package userevents

import (
	"encoding/json"
	"log"
	"server/internal/events"
	"server/internal/websocket"
)

type ReactionEvent struct {
	ReactionType string `json:"reaction_type"`
}

type ReactionHandler struct {
	connPool *websocket.WSConnectionPool
}

func NewReactionHandler(connPool *websocket.WSConnectionPool) *ReactionHandler {
	return &ReactionHandler{
		connPool: connPool,
	}
}

func (h *ReactionHandler) Handler(roomID, senderID string, data json.RawMessage) error {
	var reactionData ReactionEvent
	if err := json.Unmarshal(data, &reactionData); err != nil {
		return err
	}

	event := events.BaseEvent{
		RoomID:   roomID,
		SenderID: senderID,
		Type:     events.ReactionSent,
		Data:     data,
	}

	if err := json.Unmarshal(data, &event); err != nil {
		return err
	}

	log.Printf("#### broadcasting to all: Type=%s, Data=%s", event.Type, string(event.Data))
	h.connPool.BroadcastToAll(roomID, event)
	return nil
}

func (h *ReactionHandler) GetEventType() string {
	return "reaction"
}
