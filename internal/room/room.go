package room

import (
	"server/internal/participant"
	"server/internal/rdb"
	"strconv"
)

func GetParticipant(roomID, participantID string) (*participant.Participant, error) {
	participantData, err := rdb.Client().HGetAll(rdb.Context(), "room:"+roomID+":participant:"+participantID).Result()
	if err != nil {
		return nil, err
	}

	isHost, err := strconv.ParseBool(participantData["isHost"])
	if err != nil {
		return nil, err
	}

	participant := &participant.Participant{
		ID:        participantData["id"],
		Username:  participantData["username"],
		IsHost:    isHost,
		AvatarSrc: participantData["avatar_src"],
		SessionID: participantData["session_id"],
	}

	return participant, nil
}

func GetCallParticipants(roomID string) ([]*participant.Participant, error) {
	participantsID, err := rdb.Client().SMembers(rdb.Context(), "room:"+roomID+":participants").Result()
	if err != nil {
		return nil, err
	}

	participants := make([]*participant.Participant, 0, len(participantsID))
	for _, participantID := range participantsID {
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

		participant := &participant.Participant{
			ID:        participantData["id"],
			Username:  participantData["username"],
			IsHost:    isHost,
			AvatarSrc: participantData["avatar_src"],
			SessionID: participantData["session_id"],
		}

		participants = append(participants, participant)
	}

	return participants, err
}

func GetHost(roomID string) (string, error) {
	hostID, err := rdb.Client().HGet(rdb.Context(), "room:"+roomID, "host_id").Result()
	if err != nil {
		return "", err
	}

	return hostID, err
}

func GetRoom(id string) (string, error) {
	roomData, err := rdb.Client().HGetAll(rdb.Context(), "room:"+id).Result()
	if err != nil || len(roomData) == 0 {
		return "", err
	}

	return roomData["id"], nil
}
