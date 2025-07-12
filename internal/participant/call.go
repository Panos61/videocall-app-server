package participant

import (
	"server/internal/rdb"
)

// Sets participant's call data in Redis (username, avatar_src)
func SetParticipantCallData(roomID, participantID, username, avatarSrc string) (*Participant, error) {
	_, err := rdb.Client().HMSet(rdb.Context(), "room:"+roomID+":participant:"+participantID, map[string]any{
		"username":   username,
		"avatar_src": avatarSrc,
	}).Result()

	if err != nil {
		return nil, err
	}

	p := &Participant{
		ID:        participantID,
		Username:  username,
		AvatarSrc: avatarSrc,
	}

	return p, nil
}
