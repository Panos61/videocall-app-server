package systemevents

import (
	"server/internal/events"
	"time"
)

type HostLeftPayload struct {
	PreviousHostID string    `json:"previous_host_id"`
	Timestamp      time.Time `json:"timestamp"`
}

type HostUpdatedPayload struct {
	NewHostID string    `json:"new_host_id"`
	Timestamp time.Time `json:"timestamp"`
}

func handleHostLeft(roomID string, payload HostLeftPayload) {
	PublishSystemEvent(roomID, SystemEvent{
		Type: events.HostLeft,
		Payload: map[string]any{
			"previous_host_id": payload.PreviousHostID,
			"timestamp":        payload.Timestamp,
		},
	})
}

func handleHostUpdated(roomID string, payload HostUpdatedPayload) {
	PublishSystemEvent(roomID, SystemEvent{
		Type: events.HostUpdated,
		Payload: map[string]any{
			"new_host_id": payload.NewHostID,
			"timestamp":   payload.Timestamp,
		},
	})
}
