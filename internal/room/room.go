package room

import (
	"server/internal/rdb"
	"strconv"
	"time"
)

type RoomInfo struct {
	ID        string    `json:"id"`
	HostID    string    `json:"host_id"`
	CreatedAt time.Time `json:"created_at"`
}

func GetInfo(roomID string) (RoomInfo, error) {
	roomData, err := rdb.Client().HGetAll(rdb.Context(), "room:"+roomID).Result()
	if err != nil {
		return RoomInfo{}, err
	}

	createdAtUnix, _ := strconv.ParseInt(roomData["created_at"], 10, 64)
	createdAt := time.Unix(createdAtUnix, 0)

	return RoomInfo{ID: roomData["id"], HostID: roomData["host_id"], CreatedAt: createdAt}, nil
}

func GetRoomID(id string) (string, error) {
	roomID, err := rdb.Client().HGet(rdb.Context(), "room:"+id, "id").Result()
	if err != nil {
		return "", err
	}

	return roomID, nil
}

func SetRoomLeader(roomID, leaderID string) error {
	return rdb.Client().HSet(rdb.Context(), "room:"+roomID, "leader_id", leaderID).Err()
}

func GetRoomLeader(roomID string) (string, error) {
	leaderId, err := rdb.Client().HGet(rdb.Context(), "room:"+roomID, "leader_id").Result()
	if err != nil {
		return "", err
	}

	return leaderId, nil
}

func IsRoomLeader(roomID, participantID string) bool {
	leaderID, err := GetRoomLeader(roomID)
	if err != nil {
		return false
	}

	return leaderID == participantID
}
