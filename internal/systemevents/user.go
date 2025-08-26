package systemevents

import (
	"encoding/json"
	"server/internal/events"
	"server/internal/participant"
	"server/internal/rdb"
)

func UserJoinedEvent(roomID string, p *participant.Participant) (bool, error) {
	payload := SystemEvent{
		Type: events.UserJoined,
		Payload: map[string]any{
			"participant_id":   p.ID,
			"participant_name": p.Username,
		},
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return false, err
	}

	if err := rdb.Client().Publish(rdb.Context(), "room:"+roomID+":system_events", payloadJSON).Err(); err != nil {
		return false, err
	}

	return true, nil
}
