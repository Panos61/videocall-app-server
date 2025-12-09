package systemevents

import "encoding/json"

func handleEvent[T any](roomID string, payload map[string]any, handler func(string, T)) {
	var typedPayload T

	if jsonData, err := json.Marshal(payload); err == nil {
		if err := json.Unmarshal(jsonData, &typedPayload); err == nil {
			handler(roomID, typedPayload)
		}
	}
}
