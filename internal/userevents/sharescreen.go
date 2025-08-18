package userevents

import (
	"encoding/json"
	"fmt"
	"server/internal/events"
	"server/internal/participant"
	"server/internal/websocket"
)

type ShareScreenEvent struct {
	Username string `json:"username"`
	TrackSID string `json:"track_sid"`
	Active   bool   `json:"active"`
}

type ShareScreenHandler struct {
	connPool *websocket.WSConnectionPool
}

func NewShareScreenHandler(connPool *websocket.WSConnectionPool) *ShareScreenHandler {
	return &ShareScreenHandler{
		connPool: connPool,
	}
}

func (h *ShareScreenHandler) Handler(roomID, senderID string, payload json.RawMessage) error {
	var shareScreenData ShareScreenEvent
	if err := json.Unmarshal(payload, &shareScreenData); err != nil {
		return err
	}

	participantData, err := participant.GetParticipant(roomID, senderID)
	if err != nil {
		return err
	}

	payload, err = json.Marshal(ShareScreenEvent{
		Username: participantData.Username,
		TrackSID: shareScreenData.TrackSID,
		Active:   shareScreenData.Active,
	})
	if err != nil {
		return err
	}

	fmt.Println("shareScreenData --->>", string(payload))
	fmt.Println("shareScreenData.Active --->>", shareScreenData.Active)

	if shareScreenData.Active {
		event := events.BaseEvent{
			RoomID:   roomID,
			SenderID: senderID,
			Type:     events.ShareScreenStarted,
			Payload:  payload,
		}

		h.connPool.BroadcastToAll(roomID, event)
	}

	if !shareScreenData.Active {
		event := events.BaseEvent{
			RoomID:   roomID,
			SenderID: senderID,
			Type:     events.ShareScreenEnded,
			Payload:  payload,
		}

		h.connPool.BroadcastToAll(roomID, event)
	}

	return nil
}

func (h *ShareScreenHandler) GetEventType() string {
	return "share_screen.started"
}
