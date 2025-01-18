package room

import (
	"server/internal/participant"
	"server/internal/rdb"
)

func StartCall(roomID, participantID, username, avatarSrc string) (*participant.Participant, error) {
	_, err := rdb.Client().HMSet(rdb.Context(), "room:"+roomID+":participant:"+participantID, map[string]interface{}{
		"username":   username,
		"avatar_src": avatarSrc,
	}).Result()

	if err != nil {
		return nil, err
	}

	p := &participant.Participant{
		ID:        participantID,
		Username:  username,
		AvatarSrc: avatarSrc,
	}

	return p, nil
}
