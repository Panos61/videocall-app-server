package call

import (
	"encoding/json"
	"server/internal/participant"
	"server/internal/rdb"
	"server/internal/room"
	"strconv"
	"time"
)

type CallState struct {
	RoomID    string    `json:"room_id"`
	IsActive  bool      `json:"is_active"`
	StartedBy string    `json:"started_by"`
	StartedAt time.Time `json:"started_at"`
}

func StartCall(roomID, participantID string) (CallState, error) {
	callState := CallState{
		RoomID:    roomID,
		IsActive:  true,
		StartedBy: participantID,
		StartedAt: time.Now(),
	}

	callStateJSON, err := json.Marshal(callState)
	if err != nil {
		return CallState{}, err
	}

	err = rdb.Client().HSet(rdb.Context(), "call:"+roomID, map[string]any{
		"room_id":    roomID,
		"is_active":  true,
		"started_by": participantID,
		"started_at": time.Now().Unix(),
	}).Err()
	if err != nil {
		return CallState{}, err
	}

	if err = rdb.Client().Publish(rdb.Context(), "room:"+roomID+":call", callStateJSON).Err(); err != nil {
		return CallState{}, err
	}

	leaderID, err := room.GetRoomLeader(roomID)
	if err != nil || leaderID == "" {
		room.SetRoomLeader(roomID, participantID)
	}

	return callState, nil
}

func LeaveCall(roomID, participantID string) (bool, error) {
	participantData, err := participant.GetParticipant(roomID, participantID)
	if err != nil {
		return false, nil
	}

	_, participantsInCall, err := participant.GetParticipants(roomID)
	if err != nil {
		return false, err
	}

	if room.IsRoomLeader(roomID, participantID) {
		if len(participantsInCall) > 1 {
			room.SetRoomLeader(roomID, participantsInCall[0].ID)
		}
	}

	pipe := rdb.Client().TxPipeline()
	pipe.Del(rdb.Context(), "session:"+participantData.SessionID)
	pipe.HDel(rdb.Context(), "room:"+roomID+":participant:"+participantData.ID, "session_id")

	if len(participantsInCall) == 1 {
		pipe.Del(rdb.Context(), "call:"+roomID)
	}

	_, err = pipe.Exec(rdb.Context())
	if err != nil {
		return false, err
	}

	return true, nil
}

func GetCallState(roomID string) (CallState, error) {
	callData, err := rdb.Client().HGetAll(rdb.Context(), "call:"+roomID).Result()
	if err != nil {
		return CallState{}, err
	}

	if len(callData) == 0 {
		return CallState{
			RoomID:    roomID,
			IsActive:  false,
			StartedBy: "",
			StartedAt: time.Time{},
		}, nil
	}

	isActive, _ := strconv.ParseBool(callData["is_active"])

	startedAtUnix, _ := strconv.ParseInt(callData["started_at"], 10, 64)
	startedAt := time.Unix(startedAtUnix, 0)

	return CallState{
		RoomID:    callData["room_id"],
		IsActive:  isActive,
		StartedBy: callData["started_by"],
		StartedAt: startedAt,
	}, nil
}
