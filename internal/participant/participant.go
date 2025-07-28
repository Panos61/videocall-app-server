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

func GetParticipants(roomID string) ([]*Participant, []*Participant, error) {
	luaScript := `
		local roomID = ARGV[1]
		local participantIDs = redis.call('SMEMBERS', 'room:' .. roomID .. ':participants')
		local result = {}
		
		for i = 1, #participantIDs do
			local participantData = redis.call('HMGET', 
				'room:' .. roomID .. ':participant:' .. participantIDs[i],
				'id', 'username', 'isHost', 'avatar_src', 'session_id'
			)
			
			-- Only include participants with username
			if participantData[2] and participantData[2] ~= '' then
				table.insert(result, {
					participantData[1], -- id
					participantData[2], -- username  
					participantData[3], -- isHost
					participantData[4], -- avatar_src
					participantData[5]  -- session_id
				})
			end
		end
		
		return result
	`

	result, err := rdb.Client().Eval(rdb.Context(), luaScript, []string{}, roomID).Result()
	if err != nil {
		return []*Participant{}, []*Participant{}, err
	}

	participantList, ok := result.([]any)
	if !ok {
		return []*Participant{}, []*Participant{}, nil
	}

	allParticipants := make([]*Participant, 0, len(participantList))
	for _, item := range participantList {
		data, ok := item.([]any)
		if !ok || len(data) != 5 {
			continue
		}

		id, _ := data[0].(string)
		username, _ := data[1].(string)
		isHostStr, _ := data[2].(string)
		avatarSrc, _ := data[3].(string)
		sessionID := ""
		if data[4] != nil {
			sessionID, _ = data[4].(string)
		}

		isHost, err := strconv.ParseBool(isHostStr)
		if err != nil {
			continue
		}

		participant := &Participant{
			ID:        id,
			Username:  username,
			IsHost:    isHost,
			AvatarSrc: avatarSrc,
			SessionID: sessionID,
		}

		allParticipants = append(allParticipants, participant)
	}

	// participants = ALL participants in room (with username)
	participants := make([]*Participant, 0, len(allParticipants))
	// participantsInCall = ONLY participants actively in call (with session_id)
	participantsInCall := make([]*Participant, 0, len(allParticipants))

	for _, participant := range allParticipants {
		participants = append(participants, participant)

		if participant.SessionID != "" {
			participantsInCall = append(participantsInCall, participant)
		}
	}

	return participants, participantsInCall, nil
}
