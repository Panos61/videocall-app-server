package media

import (
	"encoding/json"
	"fmt"
	"server/internal/rdb"
)

type MediaState struct {
	Video bool `json:"video"`
	Audio bool `json:"audio"`
}

func UpdateMedia(roomID, participantID string, mediaState MediaState) (*MediaState, error) {
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
