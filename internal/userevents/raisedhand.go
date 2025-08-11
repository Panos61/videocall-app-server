package userevents

import (
	"encoding/json"
	"server/internal/events"
	"server/internal/participant"
	"server/internal/websocket"
)

type RaisedHandEvent struct {
	HandRaised bool   `json:"raised_hand"`
	Username   string `json:"username"`
}

type RaisedHandHandler struct {
	connPool *websocket.WSConnectionPool
}

func NewRaisedHandHandler(connPool *websocket.WSConnectionPool) *RaisedHandHandler {
	return &RaisedHandHandler{
		connPool: connPool,
	}
}

func (h *RaisedHandHandler) Handler(roomID, senderID string, payload json.RawMessage) error {
	var raisedHandEvent RaisedHandEvent
	if err := json.Unmarshal(payload, &raisedHandEvent); err != nil {
		return err
	}

	participantData, err := participant.GetParticipant(roomID, senderID)
	if err != nil {
		return err
	}

	payload, err = json.Marshal(RaisedHandEvent{
		HandRaised: raisedHandEvent.HandRaised,
		Username:   participantData.Username,
	})

	if err != nil {
		return err
	}
	event := events.BaseEvent{
		RoomID:   roomID,
		SenderID: senderID,
		Type:     events.RaisedHand,
		Payload:  payload,
	}

	h.connPool.BroadcastToAll(roomID, event)
	return nil
}

func (h *RaisedHandHandler) GetEventType() string {
	return "raised_hand.sent"
}
