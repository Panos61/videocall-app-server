package userevents

import (
	"encoding/json"
	"server/internal/events"
	"server/internal/websocket"
)

type RaisedHandEvent struct {
	HandRaised bool `json:"hand_raised"`
}

type RaisedHandHandler struct {
	connPool *websocket.WSConnectionPool
}

func NewRaisedHandHandler(connPool *websocket.WSConnectionPool) *RaisedHandHandler {
	return &RaisedHandHandler{
		connPool: connPool,
	}
}

func (h *RaisedHandHandler) Handler(roomID, senderID string, data json.RawMessage) error {
	var raisedHandEvent RaisedHandEvent
	if err := json.Unmarshal(data, &raisedHandEvent); err != nil {
		return err
	}

	event := events.BaseEvent{
		RoomID:   roomID,
		SenderID: senderID,
		Type:     events.RaisedHand,
		Data:     data,
	}

	h.connPool.BroadcastToAll(roomID, event)
	return nil
}

func (h *RaisedHandHandler) GetEventType() string {
	return "raised_hand"
}
