package systemevents

import (
	"server/internal/events"
	"time"
)

type HostLeftPayload struct {
	PreviousHostID string    `json:"previous_host_id"`
	Timestamp      time.Time `json:"timestamp"`
}

type HostHandoverPayload struct {
	PreviousCandidateID string    `json:"previous_candidate_id"`
	HasAccepted         bool      `json:"has_accepted"`
	Timestamp           time.Time `json:"timestamp"`
}

type HostUpdatePayload struct {
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

func handleHostHandover(roomID string, payload HostHandoverPayload) {
	PublishSystemEvent(roomID, SystemEvent{
		Type: events.HostHandover,
		Payload: map[string]any{
			"previous_candidate_id": payload.PreviousCandidateID,
			"has_accepted":          payload.HasAccepted,
			"timestamp":             payload.Timestamp,
		},
	})
}

func handleHostUpdate(roomID string, payload HostUpdatePayload) {
	PublishSystemEvent(roomID, SystemEvent{
		Type: events.HostUpdated,
		Payload: map[string]any{
			"new_host_id": payload.NewHostID,
			"timestamp":   payload.Timestamp,
		},
	})
}
