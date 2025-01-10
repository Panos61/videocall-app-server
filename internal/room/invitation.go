package room

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/url"
	"server/internal/rdb"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

func GenerateInvitationCode(roomID string) string {
	buffer := make([]byte, 12)

	_, err := rand.Read(buffer)
	if err != nil {
		log.Printf("Error generating invitation code: %s\n", err)
		return ""
	}

	invitationCode := base64.URLEncoding.EncodeToString(buffer)
	return invitationCode
}

func BuildInvitationURL(roomID, invitationCode string) string {
	invitationURL := url.URL{
		Scheme: "http",
		Host:   "localhost:5173",
		Path:   "/room-invite",
		RawQuery: url.Values{
			"room": {roomID},
			"code": {invitationCode},
		}.Encode(),
	}

	return invitationURL.String()
}

func SetExpiration(roomID string, inviteExpiration string) (bool, error) {
	expiryToStr, err := strconv.Atoi(inviteExpiration)
	if err != nil {
		return false, err
	}

	expiresIn := time.Duration(expiryToStr) * time.Minute

	err = rdb.Client().HSet(rdb.Context(), "room:"+roomID, "expiresIn", time.Now().Add(expiresIn).Format(time.RFC3339)).Err()
	if err != nil {
		return false, err
	}

	return true, nil
}

func GetExpiration(roomID string) (time.Duration, error) {
	defaultExpiration := 30 * time.Minute

	expiresIn, err := rdb.Client().HGet(rdb.Context(), "room:"+roomID, "").Result()
	if err != nil || expiresIn == "" {
		return defaultExpiration, fmt.Errorf("failed to get expiration time %v", err)
	}

	duration, err := strconv.Atoi(expiresIn)
	if err != nil {
		return defaultExpiration, err
	}

	return time.Duration(duration) * time.Minute, nil
}

// Sets the invitation URL and expiration time for the room
func SetInvitation(roomID, invitationCode string) (string, error) {
	expirationTime, err := GetExpiration(roomID)
	if err != nil {
		// if error, get the default expiration time
		expirationTime = 30 * time.Minute
	}

	invitationURL := BuildInvitationURL(roomID, invitationCode)

	_, err = rdb.Client().HSet(rdb.Context(), "room:"+roomID, map[string]interface{}{
		"invitation": invitationURL,
		"expiresIn":  time.Now().Add(expirationTime).Format(time.RFC3339),
	}).Result()

	if err != nil {
		log.Printf("Error setting room invitation key %s\n", err)
		return "", err
	}

	return invitationURL, nil
}

func IsInvitationExpired(roomID string) (bool, error) {
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

// Checks for any existing room using the invKey reverse index mapped to roomID
func ValidateInvitation(urlInput string) (bool, string, error) {
	roomID, err := rdb.Client().Get(rdb.Context(), "invitation:"+urlInput).Result()
	if err != nil {
		if err == redis.Nil {
			return false, "", err
		} else {
			return false, "", err
		}
	}

	return true, roomID, nil
}
