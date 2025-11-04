package room

import (
	"fmt"
	"server/internal/participant"
	"server/internal/rdb"
	"server/internal/utils"
	"time"

	"github.com/google/uuid"
)

func CreateRoom() (string, error) {
	roomID := uuid.New().String()
	invitationCode := GenerateCode()

	pipe := rdb.Client().TxPipeline()
	pipe.HSet(rdb.Context(), "room:"+roomID, map[string]any{"id": roomID, "created_at": time.Now().Unix()})
	pipe.Set(rdb.Context(), "room:"+roomID+":invitation", invitationCode, 30*time.Minute)
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
		participant.BroadcastParticipantsUpdate(roomID, []*participant.Participant{
			{
				ID: p.ID,
			},
		})
	}

	return p, nil
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
		hostUpdated, err := AssignRandomHost(roomID, participantID, participantIDs)
		fmt.Println("hostUpdated", hostUpdated)
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

func KillRoom(roomID, participantID string) error {
	pipe := rdb.Client().TxPipeline()
	pipe.Del(rdb.Context(), "room:"+roomID)
	pipe.Del(rdb.Context(), "room:"+roomID+":settings")
	pipe.Del(rdb.Context(), "room:"+roomID+":invitation")
	pipe.Del(rdb.Context(), "room:"+roomID+":participant:"+participantID)
	pipe.Del(rdb.Context(), "room:"+roomID+":participants")
	pipe.Del(rdb.Context(), "call:"+roomID)

	_, err := pipe.Exec(rdb.Context())
	if err != nil {
		return err
	}

	return nil
}
