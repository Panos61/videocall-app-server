package participant

import (
	"fmt"
	"server/internal/rdb"
	"strconv"
)

type Participant struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	IsHost    bool   `json:"isHost"`
	AvatarSrc string `json:"avatar_src"`
	Token     string `json:"jwt,omitempty"`
	SessionID string `json:"session_id,omitempty"`
}

func GetMe(roomID, participantID string) (*Participant, error) {
	me, err := rdb.Client().HGetAll(rdb.Context(), "room:"+roomID+":participant:"+participantID).Result()
	if err != nil {
		return nil, err
	}

	if len(me) == 0 {
		return nil, fmt.Errorf("participant not found")
	}

	isHost, err := strconv.ParseBool(me["isHost"])
	if err != nil {
		return nil, err
	}

	participant := Participant{
		ID:        participantID,
		Username:  me["username"],
		IsHost:    isHost,
		AvatarSrc: me["avatar_src"],
	}

	return &participant, nil
}
