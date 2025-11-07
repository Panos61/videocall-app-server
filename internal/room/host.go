package room

import (
	"fmt"
	"math/rand"
	"server/internal/participant"
	"server/internal/rdb"
	"server/internal/utils"
	"time"
)

// sets as host the creator of the room
func SetHostParticipant(roomID string) (*participant.Participant, error) {
	participantID := utils.GenerateParticipantID()

	pipe := rdb.Client().TxPipeline()
	pipe.SAdd(rdb.Context(), "room:"+roomID+":participants", participantID)
	pipe.HMSet(rdb.Context(), "room:"+roomID+":participant:"+participantID, map[string]any{
		"id":     participantID,
		"isHost": true,
	})
	pipe.HSet(rdb.Context(), "room:"+roomID, "host_id", participantID)

	_, err := pipe.Exec(rdb.Context())
	if err != nil {
		return nil, fmt.Errorf("error setting host participant: %w", err)
	}

	jwtToken, err := utils.GenerateJWT(participantID, true)
	if err != nil {
		return nil, err
	}

	return &participant.Participant{
		ID:     participantID,
		Token:  jwtToken,
		IsHost: true,
	}, nil
}

// assigns a random host from the non-host participants after the previous host leaves the room
func AssignRandomHost(roomID, previousHostID string, participantIDs []string) (bool, string, error) {
	var nonHostParticipants []string

	for _, id := range participantIDs {
		if id != previousHostID {
			nonHostParticipants = append(nonHostParticipants, id)
		}
	}

	if len(nonHostParticipants) == 0 {
		return false, "", nil
	}

	rand.New(rand.NewSource(time.Now().UnixNano()))
	randomIndex := rand.Intn(len(nonHostParticipants))
	newHostID := nonHostParticipants[randomIndex]

	pipe := rdb.Client().Pipeline()
	pipe.HSet(rdb.Context(), "room:"+roomID, map[string]string{"host_id": newHostID})
	pipe.HSet(rdb.Context(), "room:"+roomID+":participant:"+newHostID, map[string]any{"isHost": true})

	_, err := pipe.Exec(rdb.Context())
	if err != nil {
		return false, "", err
	}

	return true, newHostID, nil
}
