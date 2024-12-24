package room

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"server/internal/rdb"
	"time"

	"github.com/redis/go-redis/v9"
)

func GenerateInvKey(roomID string) string {
	buffer := make([]byte, 12)

	_, err := rand.Read(buffer)
	if err != nil {
		log.Printf("Error generating random bytes: %s\n", err)
		return ""
	}

	invitationKey := base64.URLEncoding.EncodeToString(buffer)
	return invitationKey
}

func SetRoomKey(roomID string, invKey string) error {
	expirationTime := 10 * time.Second

	err := rdb.Client().HSet(rdb.Context(), "room:"+roomID, map[string]interface{}{
		"invitation_key": invKey,
		"expiresIn":      time.Now().Add(expirationTime).Format(time.RFC3339),
	}).Err()

	if err != nil {
		log.Printf("Error setting room invitation key %s\n", err)
		return err
	}

	return nil
}

func IsKeyExpired(roomID string) (bool, error) {
	expiresIn, err := rdb.Client().HGet(rdb.Context(), "room:"+roomID, "expiresIn").Result()
	if err != nil {
		return true, nil
	}

	expirationTime, err := time.Parse(time.RFC3339, expiresIn)
	if err != nil {
		return false, fmt.Errorf("failed to parse expiry time: %v", err)
	}

	return time.Now().After(expirationTime), nil
}

func GetCurrentKey(roomID string) (string, error) {
	key, err := rdb.Client().HGet(rdb.Context(), roomID, "invitation_key").Result()

	if key == "" || err != nil {
		return "", err
	}

	return key, nil
}

func InvitationKeyReverseIndex(invitationKey, roomID string) error {
	// Creates a reverse index mapping invitationKey to roomID
	expires := 10 * time.Second

	err := rdb.Client().Set(rdb.Context(), "invitationKey:"+invitationKey, roomID, expires).Err()
	if err != nil {
		return fmt.Errorf("failed to create invkey reverse index %w", err)
	}

	return nil
}

func AuthorizeInvitationKey(keyInput string) (bool, string, error) {
	// Checks for any existing room using the invKey reverse index mapped to roomID
	roomID, err := rdb.Client().Get(rdb.Context(), "invitationKey:"+keyInput).Result()
	if err != nil {
		if err == redis.Nil {
			return false, "", err
		} else {
			return false, "", err
		}
	}

	return true, roomID, nil
}
