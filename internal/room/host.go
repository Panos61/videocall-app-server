package room

import (
	"fmt"
	"math/rand"
	"server/internal/events"
	"server/internal/participant"
	"server/internal/rdb"
	"server/internal/systemevents"
	"server/internal/utils"
	"time"
)

// sets as host the creator of the room
func SetCreatorAsHost(roomID string) (*participant.Participant, error) {
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

	jwtToken, err := utils.GenerateJWT(participantID)
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

// manually sets the selected participant as the new host
// the previous host is unassigned from host
func SetHost(roomID, currentHostID, selectedParticipantID string) error {
	participant, err := participant.GetParticipant(roomID, selectedParticipantID)
	if err != nil {
		return fmt.Errorf("selected participant does not exist: %v", err)
	}

	if participant.IsHost {
		return fmt.Errorf("selected participant is already host")
	}

	pipe := rdb.Client().TxPipeline()
	pipe.HSet(rdb.Context(), "room:"+roomID, "host_id", selectedParticipantID)
	pipe.HSet(rdb.Context(), "room:"+roomID+":participant:"+selectedParticipantID, map[string]any{"isHost": true})
	pipe.HSet(rdb.Context(), "room:"+roomID+":participant:"+currentHostID, map[string]any{"isHost": false})

	_, err = pipe.Exec(rdb.Context())
	if err != nil {
		return err
	}

	systemevents.PublishSystemEvent(roomID, systemevents.SystemEvent{
		Type: events.HostUpdated,
		Payload: map[string]any{
			"current_host_id": currentHostID,
			"new_host_id":     selectedParticipantID,
			"timestamp":       time.Now().Unix(),
		},
	})

	return nil
}

func GetHost(roomID string) (string, error) {
	hostID, err := rdb.Client().HGet(rdb.Context(), "room:"+roomID, "host_id").Result()
	if err != nil {
		return "", err
	}

	return hostID, err
}
