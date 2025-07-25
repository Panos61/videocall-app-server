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

func GetParticipant(roomID, participantID string) (*Participant, error) {
	participantData, err := rdb.Client().HGetAll(rdb.Context(), "room:"+roomID+":participant:"+participantID).Result()
	if err != nil {
		return nil, err
	}

	isHost, err := strconv.ParseBool(participantData["isHost"])
	if err != nil {
		return nil, err
	}

	participant := &Participant{
		ID:        participantData["id"],
		Username:  participantData["username"],
		IsHost:    isHost,
		AvatarSrc: participantData["avatar_src"],
		SessionID: participantData["session_id"],
	}

	return participant, nil
}

// todo: fix N+1
func GetCallParticipants(roomID string) ([]*Participant, error) {
	participantsID, err := rdb.Client().SMembers(rdb.Context(), "room:"+roomID+":participants").Result()
	if err != nil {
		return nil, err
	}

	participants := make([]*Participant, 0, len(participantsID))
	for _, participantID := range participantsID {
		fmt.Println("participantID", participantID)
		participantData, err := rdb.Client().HGetAll(rdb.Context(), "room:"+roomID+":participant:"+participantID).Result()
		if err != nil {
			return nil, err
		}

		// skip any participant with no session id
		if len(participantData["session_id"]) == 0 {
			continue
		}

		isHost, err := strconv.ParseBool(participantData["isHost"])
		if err != nil {
			return nil, err
		}

		participant := &Participant{
			ID:        participantData["id"],
			Username:  participantData["username"],
			IsHost:    isHost,
			AvatarSrc: participantData["avatar_src"],
			SessionID: participantData["session_id"],
		}

		participants = append(participants, participant)
		fmt.Println(participants)
	}

	return participants, err
}
