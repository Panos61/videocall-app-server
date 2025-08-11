package events

import (
	"encoding/json"
	"errors"
)

type BaseEvent struct {
	RoomID   string          `json:"room_id"`
	SenderID string          `json:"sender_id"`
	Type     string          `json:"type"`
	Payload  json.RawMessage `json:"payload"`
}

type EventHandler interface {
	Handler(roomID, senderID string, data json.RawMessage) error
	GetEventType() string
}

type EventRegistry struct {
	handlers map[string]EventHandler
}

func NewEventRegistry() *EventRegistry {
	return &EventRegistry{
		handlers: make(map[string]EventHandler),
	}
}

func (r *EventRegistry) RegisterHandler(handler EventHandler) {
	r.handlers[handler.GetEventType()] = handler
}

func (r *EventRegistry) HandleEvent(event BaseEvent) error {
	handler, exists := r.handlers[event.Type]
	if !exists {
		return errors.New("event handler not found")
	}

	return handler.Handler(event.RoomID, event.SenderID, event.Payload)
}
