package room

import (
	"server/internal/rdb"
	"strconv"
	"time"
)

type RoomInfo struct {
	CreatedAt time.Time `json:"created_at"`
}

func GetInfo(roomID string) (RoomInfo, error) {
	roomData, err := rdb.Client().HGetAll(rdb.Context(), "room:"+roomID).Result()
	if err != nil {
		return RoomInfo{}, err
	}

	createdAtUnix, _ := strconv.ParseInt(roomData["created_at"], 10, 64)
	createdAt := time.Unix(createdAtUnix, 0)

	return RoomInfo{CreatedAt: createdAt}, nil
}

func GetRoom(id string) (string, error) {
	roomData, err := rdb.Client().HGetAll(rdb.Context(), "room:"+id).Result()
	if err != nil || len(roomData) == 0 {
		return "", err
	}

	return roomData["id"], nil
}
