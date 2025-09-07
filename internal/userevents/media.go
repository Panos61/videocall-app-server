package userevents

import (
	"encoding/json"
	"server/internal/events"
	"server/internal/websocket"
)

type MediaControlEvent struct {
	AudioEnabled bool `json:"audio"`
	VideoEnabled bool `json:"video"`
}

type MediaControlHandler struct {
	connPool *websocket.WSConnectionPool
}

func NewMediaControlHandler(connPool *websocket.WSConnectionPool) *MediaControlHandler {
	return &MediaControlHandler{
		connPool: connPool,
	}
}

func (h *MediaControlHandler) Handler(roomID, senderID, sessionID string, payload json.RawMessage) error {
	var mediaControlEvent MediaControlEvent
	if err := json.Unmarshal(payload, &mediaControlEvent); err != nil {
		return err
	}

	payload, err := json.Marshal(mediaControlEvent)
	if err != nil {
		return err
	}

	event := events.BaseEvent{
		RoomID:    roomID,
		SessionID: sessionID,
		Type:      events.MediaControlUpdated,
		Payload:   payload,
	}

	h.connPool.BroadcastToAll(roomID, event)
	return nil
}

func (h *MediaControlHandler) GetEventType() string {
	return "media.control.updated"
}

type SyncMediaStateEvent map[string]MediaControlEvent

type SyncMediaHandler struct {
	connPool *websocket.WSConnectionPool
}

func NewSyncMediaHandler(connPool *websocket.WSConnectionPool) *SyncMediaHandler {
	return &SyncMediaHandler{
		connPool: connPool,
	}
}

func (h *SyncMediaHandler) Handler(roomID, senderID, sessionID string, payload json.RawMessage) error {
	if !h.connPool.IsRoomLeader(roomID, senderID) {
		return nil
	}

	var syncMediaEvent SyncMediaStateEvent
	if err := json.Unmarshal(payload, &syncMediaEvent); err != nil {
		return err
	}

	payload, err := json.Marshal(syncMediaEvent)
	if err != nil {
		return err
	}

	event := events.BaseEvent{
		RoomID:    roomID,
		SessionID: sessionID,
		Type:      events.SyncMedia,
		Payload:   payload,
	}

	h.connPool.BroadcastToAll(roomID, event)
	return nil
}

func (h *SyncMediaHandler) GetEventType() string {
	return "sync.media"
}
