package room

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"server/internal/participant"
	"server/internal/rdb"
	"server/internal/utils"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type Room struct {
	ID           string                              `json:"id"`
	Participants map[string]*participant.Participant `json:"participants"`
	HostID       string                              `json:"host_id"`
}

func CreateRoom() (*Room, error) {
	roomID := uuid.New().String()

	newRoom := &Room{
		ID:           roomID,
		Participants: make(map[string]*participant.Participant),
	}

	pipe := rdb.Client().TxPipeline()
	pipe.HSet(rdb.Context(), "room:"+roomID, map[string]interface{}{"id": roomID})
	pipe.HSet(rdb.Context(), "room:"+roomID+":settings", map[string]interface{}{
		"invitation_expiry": "30",
		"invite_permission": false,
	})

	_, err := pipe.Exec(rdb.Context())
	if err != nil {
		return nil, fmt.Errorf("error creating room: %w", err)
	}

	return newRoom, nil
}

func JoinRoom(roomID string) (*participant.Participant, error) {
	participant := &participant.Participant{
		ID:       utils.GenerateParticipantID(),
		Username: "",
		IsHost:   false,
	}

	media := map[string]bool{
		"audio": false,
		"video": false,
	}

	mediaJSON, err := json.Marshal(media)
	if err != nil {
		return nil, err
	}

	pipe := rdb.Client().TxPipeline()
	pipe.SAdd(rdb.Context(), "room:"+roomID+":participants", participant.ID)
	pipe.HMSet(rdb.Context(), "room:"+roomID+":participant:"+participant.ID, map[string]interface{}{
		"id":       participant.ID,
		"username": participant.Username,
		"isHost":   participant.IsHost,
		"media":    string(mediaJSON),
	})

	_, err = pipe.Exec(rdb.Context())
	if err != nil {
		return nil, fmt.Errorf("error setting host participant: %w", err)
	}

	token, err := utils.GenerateJWT(participant.ID, false)
	if err != nil {
		return nil, fmt.Errorf("error generating token for guest: %w", err)
	}

	participant.Token = token
	return participant, nil
}

func SetHostParticipant(roomID string) (*participant.Participant, error) {
	participant := &participant.Participant{
		ID:     utils.GenerateParticipantID(),
		IsHost: true,
	}

	media := map[string]bool{
		"audio": false,
		"video": false,
	}

	mediaJSON, err := json.Marshal(media)
	if err != nil {
		return nil, err
	}

	pipe := rdb.Client().TxPipeline()
	pipe.SAdd(rdb.Context(), "room:"+roomID+":participants", participant.ID)
	pipe.HMSet(rdb.Context(), "room:"+roomID+":participant:"+participant.ID, map[string]interface{}{
		"id":     participant.ID,
		"isHost": participant.IsHost,
		"media":  mediaJSON,
	})
	pipe.HSet(rdb.Context(), "room:"+roomID, "host_id", participant.ID)

	_, err = pipe.Exec(rdb.Context())
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

func LeaveRoom(roomID, participantID string) (bool, error) {
	participant, err := GetParticipant(roomID, participantID)
	if err != nil {
		return false, nil
	}

	participantIDs, err := rdb.Client().SMembers(rdb.Context(), "room:"+roomID+":participants").Result()
	if err != nil {
		return false, nil
	}

	pipe := rdb.Client().TxPipeline()
	pipe.Del(rdb.Context(), "room:"+roomID+":participant:"+participantID)
	pipe.SRem(rdb.Context(), "room:"+roomID+":participants", participant.ID)
	pipe.Del(rdb.Context(), "session:"+participant.SessionID)
	if participant.IsHost {
		_, err := UpdateHost(roomID, participant.ID, participantIDs)
		if err != nil {
			return false, nil
		}
	}

	if len(participantIDs) != 1 {
		if participant.IsHost {
			_, err := notifyHostLeft(roomID, participant)
			if err != nil {
				return false, err
			}
		} else {
			_, err := notifyUserLeft(roomID, participant)
			if err != nil {
				return false, err
			}
		}
	}

	if len(participantIDs) == 1 {
		pipe.Del(rdb.Context(), "room:"+roomID)
		pipe.Del(rdb.Context(), "room:"+roomID+":participants")
	}

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
	pipe.HSet(rdb.Context(), "room:"+roomID, map[string]interface{}{"host_id": newHostID})
	pipe.HSet(rdb.Context(), "room:"+roomID+":participant:"+newHostID, map[string]interface{}{"isHost": true})

	_, err := pipe.Exec(rdb.Context())
	if err != nil {
		return false, err
	}

	return true, nil
}

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

func GetRoom(id string) (*Room, error) {
	roomData, err := rdb.Client().HGetAll(rdb.Context(), "room:"+id).Result()
	if err != nil || len(roomData) == 0 {
		return nil, err
	}

	room := &Room{
		ID: roomData["id"],
	}

	return room, nil
}
