package userevents

import (
	"encoding/json"
	"server/internal/events"
	"server/internal/participant"
	"server/internal/websocket"
)

type ReactionEvent struct {
	ReactionType string `json:"reaction_type"`
	Username     string `json:"username"`
}

type ReactionHandler struct {
	connPool *websocket.WSConnectionPool
}

func NewReactionHandler(connPool *websocket.WSConnectionPool) *ReactionHandler {
	return &ReactionHandler{
		connPool: connPool,
	}
}

func (h *ReactionHandler) Handler(roomID, senderID string, payload json.RawMessage) error {
	var reactionData ReactionEvent
	if err := json.Unmarshal(payload, &reactionData); err != nil {
		return err
	}

	participantData, err := participant.GetParticipant(roomID, senderID)
	if err != nil {
		return err
	}

	payload, err = json.Marshal(ReactionEvent{
		ReactionType: reactionData.ReactionType,
		Username:     participantData.Username,
	})
	if err != nil {
		return err
	}

	event := events.BaseEvent{
		RoomID:   roomID,
		SenderID: senderID,
		Type:     events.ReactionSent,
		Payload:  payload,
	}

	h.connPool.BroadcastToAll(roomID, event)
	return nil
}

func (h *ReactionHandler) GetEventType() string {
	return "reaction.sent"
}
