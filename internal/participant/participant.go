package participant

import (
	"encoding/json"
	"fmt"
	"server/internal/rdb"
	"strconv"
	"time"
)

type Participant struct {
	ID         string     `json:"id"`
	Username   string     `json:"username"`
	IsHost     bool       `json:"isHost"`
	AvatarSrc  string     `json:"avatar_src"`
	MediaState MediaState `json:"media"`
	Token      string     `json:"jwt,omitempty"`
	SessionID  string     `json:"session_id,omitempty"`
}

type MediaState struct {
	Video bool `json:"video"`
	Audio bool `json:"audio"`
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
		MediaState: MediaState{
			Video: me["video"] == "true",
			Audio: me["audio"] == "true",
		},
	}

	return &participant, nil
}

// Check if user is authorized to join the room
func IsUserAuthorized(roomID string) (bool, error) {
	if roomID == "" {
		return false, nil
	}

	expiresIn, err := rdb.Client().HGet(rdb.Context(), "room:"+roomID, "expiresIn").Result()
	if err != nil {
		return false, err
	}

	expirationTime, err := time.Parse(time.RFC3339, expiresIn)
	if err != nil {
		return false, err
	}

	return !time.Now().After(expirationTime), nil
}

func UpdateUserMediaState(roomID, participantID string, mediaState MediaState) (*MediaState, error) {
	media := map[string]bool{
		"audio": mediaState.Audio,
		"video": mediaState.Video,
	}

	mediaJSON, err := json.Marshal(media)
	if err != nil {
		return nil, err
	}

	err = rdb.Client().HSet(rdb.Context(), "room:"+roomID+":participant:"+participantID, map[string]interface{}{
		"media": mediaJSON,
	}).Err()
	if err != nil {
		return nil, err
	}

	fmt.Printf("media %v\n", media)

	return &mediaState, nil
}
