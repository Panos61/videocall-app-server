package room

import (
	"fmt"
	"server/internal/rdb"
)

func StartCall(roomID, participantID, username, avatarSrc string) (*Participant, error) {
	var participant *Participant

	_, err := rdb.Client().HSet(rdb.Context(), "room:"+roomID+":participant:"+participantID, map[string]interface{}{
		"username":   username,
		"avatar_src": avatarSrc,
	}).Result()

	if err != nil {
		return nil, err
	}

	participant = &Participant{
		Username: username,
	}

	fmt.Printf("room/call participant: %+v\n", participant)

	return participant, nil
}
