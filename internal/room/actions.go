package room

import (
	"fmt"
	"math/rand"
	"server/internal/participant"
	"server/internal/rdb"
	"server/internal/utils"
	"time"

	"github.com/google/uuid"
)

func CreateRoom() (string, error) {
	roomID := uuid.New().String()

	code := GenerateCode()

	pipe := rdb.Client().TxPipeline()
	pipe.HSet(rdb.Context(), "room:"+roomID, map[string]any{"id": roomID, "created_at": time.Now().Unix()})
	pipe.Set(rdb.Context(), "room:"+roomID+":invitation", code, 30*time.Minute)
	pipe.HSet(rdb.Context(), "room:"+roomID+":settings", map[string]any{
		"invitation_expiry": "30",
		"invite_permission": false,
	})

	_, err := pipe.Exec(rdb.Context())
	if err != nil {
		return "", fmt.Errorf("error creating room: %w", err)
	}

	return roomID, nil
}

func JoinRoom(roomID string) (*participant.Participant, error) {
	p := &participant.Participant{
		ID:       utils.GenerateParticipantID(),
		Username: "",
		IsHost:   false,
	}

	pipe := rdb.Client().TxPipeline()
	pipe.SAdd(rdb.Context(), "room:"+roomID+":participants", p.ID)
	pipe.HMSet(rdb.Context(), "room:"+roomID+":participant:"+p.ID, map[string]any{
		"id":       p.ID,
		"username": p.Username,
		"isHost":   p.IsHost,
	})

	_, err := pipe.Exec(rdb.Context())
	if err != nil {
		return nil, fmt.Errorf("error setting host participant: %w", err)
	}

	token, err := utils.GenerateJWT(p.ID, false)
	if err != nil {
		return nil, fmt.Errorf("error generating token for guest: %w", err)
	}

	p.Token = token

	// Broadcast participants-in-lobby update (only ID, as username might not be set yet in lobby)
	if p.SessionID == "" {
		participant.BroadcastParticipantsUpdate(roomID, []participant.Guest{
			{
				ID: p.ID,
			},
		})
	}

	return p, nil
}

func SetHostParticipant(roomID string) (*participant.Participant, error) {
	participant := &participant.Participant{
		ID:     utils.GenerateParticipantID(),
		IsHost: true,
	}

	pipe := rdb.Client().TxPipeline()
	pipe.SAdd(rdb.Context(), "room:"+roomID+":participants", participant.ID)
	pipe.HMSet(rdb.Context(), "room:"+roomID+":participant:"+participant.ID, map[string]any{
		"id":     participant.ID,
		"isHost": participant.IsHost,
	})
	pipe.HSet(rdb.Context(), "room:"+roomID, "host_id", participant.ID)

	_, err := pipe.Exec(rdb.Context())
	if err != nil {
		return nil, fmt.Errorf("error setting host participant: %w", err)
	}

	jwtToken, err := utils.GenerateJWT(participant.ID, true)
	if err != nil {
		return nil, err
	}

	participant.Token = jwtToken

	return participant, nil
}

func ExitRoom(roomID, participantID string, isHost bool) (bool, error) {
	participantIDs, err := rdb.Client().SMembers(rdb.Context(), "room:"+roomID+":participants").Result()
	if err != nil {
		return false, nil
	}

	pipe := rdb.Client().TxPipeline()
	// if there's only one participant, delete the room and relevant data
	if len(participantIDs) == 1 {
		pipe.Del(rdb.Context(), "room:"+roomID)
		pipe.Del(rdb.Context(), "room:"+roomID+":settings")
		pipe.Del(rdb.Context(), "room:"+roomID+":invitation")
		pipe.Del(rdb.Context(), "room:"+roomID+":participant:"+participantID)
		pipe.Del(rdb.Context(), "room:"+roomID+":participants")

		_, err := pipe.Exec(rdb.Context())
		if err != nil {
			return false, err
		}

		return true, nil
	}

	if isHost {
		hostUpdated, err := UpdateHost(roomID, participantID, participantIDs)
		if err != nil {
			return false, err
		}

		if !hostUpdated {
			return false, nil
		}
	}

	pipe.SRem(rdb.Context(), "room:"+roomID+":participants", participantID)
	pipe.Del(rdb.Context(), "room:"+roomID+":participant:"+participantID)

	_, err = pipe.Exec(rdb.Context())
	if err != nil {
		return false, err
	}

	return true, nil
}

func UpdateHost(roomID, previousHostID string, participantIDs []string) (bool, error) {
	var nonHostParticipants []string

	for _, id := range participantIDs {
		if id != previousHostID {
			nonHostParticipants = append(nonHostParticipants, id)
		}
	}

	if len(nonHostParticipants) == 0 {
		return false, nil
	}

	rand.New(rand.NewSource(time.Now().UnixNano()))
	randomIndex := rand.Intn(len(nonHostParticipants))
	newHostID := nonHostParticipants[randomIndex]

	pipe := rdb.Client().Pipeline()
	pipe.HSet(rdb.Context(), "room:"+roomID, map[string]any{"host_id": newHostID})
	pipe.HSet(rdb.Context(), "room:"+roomID+":participant:"+newHostID, map[string]any{"isHost": true})

	_, err := pipe.Exec(rdb.Context())
	if err != nil {
		return false, err
	}

	return true, nil
}
