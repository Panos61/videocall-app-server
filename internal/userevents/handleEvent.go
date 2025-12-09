package userevents

import "encoding/json"

func handleEvent[T any](roomID, participantID string, payload map[string]any, handler func(string, string, T)) {
	var typedPayload T

	if jsonData, err := json.Marshal(payload); err == nil {
		if err := json.Unmarshal(jsonData, &typedPayload); err == nil {
			handler(roomID, participantID, typedPayload)
		}
	}
}
