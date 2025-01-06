package participant

import (
	"encoding/json"
	"fmt"
	"server/internal/rdb"
	"time"
)

type MediaState struct {
	Video bool `json:"video"`
	Audio bool `json:"audio"`
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
