package room

import "server/internal/rdb"

func StoreUserSession(roomID, sessionID, participantID string) error {
	err := rdb.Client().HMSet(rdb.Context(), "room:"+roomID+":participant:"+participantID, map[string]interface{}{
		"session_id": sessionID,
	}).Err()
	if err != nil {
		return err
	}

	// Create a reverse lookup for session:participant_id
	err = rdb.Client().Set(rdb.Context(), "session:"+sessionID, participantID, 0).Err()
	if err != nil {
		return err
	}

	return nil
}

func ValidateUserSession(sessionID string) (string, error) {
	participantID, err := rdb.Client().Get(rdb.Context(), "session:"+sessionID).Result()
	if err != nil {
		return "", err
	}

	return participantID, nil
}
