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

// Sets the invitation URL and expiration time for the room
func SetInvitation(roomID, invitationCode string) (string, error) {
	expirationTime, err := getExpirationDuration(roomID)
	if err != nil {
		// if error, get the default expiration time
		expirationTime = 30 * time.Minute
	}

	duration := strconv.Itoa(int(expirationTime.Minutes()))
	invitationURL := BuildInvitationURL(roomID, invitationCode)

	_, err = rdb.Client().HSet(rdb.Context(), "room:"+roomID, map[string]interface{}{
		"duration":   duration,
		"invitation": invitationURL,
		"expiresIn":  time.Now().Add(expirationTime).Format(time.RFC3339),
	}).Result()

	if err != nil {
		log.Printf("Error setting room invitation key %s\n", err)
		return "", err
	}

	return invitationURL, nil
}

func SetExpiration(roomID string, inviteExpiration string) (string, error) {
	expiryToStr, err := strconv.Atoi(inviteExpiration)
	if err != nil {
		return "", err
	}

	expiresIn := time.Duration(expiryToStr) * time.Minute

	err = rdb.Client().HSet(rdb.Context(), "room:"+roomID, map[string]interface{}{
		"expiresIn": time.Now().Add(expiresIn).Format(time.RFC3339),
		"duration":  inviteExpiration,
	}).Err()
	if err != nil {
		return "", err
	}

	return inviteExpiration, nil
}

// Returns the expiration duration in string format
func GetExpiry(roomID string) (string, error) {
	defaultValue := "30"

	duration, err := rdb.Client().HGet(rdb.Context(), "room:"+roomID, "duration").Result()
	if err != nil {
		return defaultValue, err
	}

	return duration, nil
}

func getExpirationDuration(roomID string) (time.Duration, error) {
	defaultExpiration := 30 * time.Minute

	durationStr, err := rdb.Client().HGet(rdb.Context(), "room:"+roomID, "duration").Result()
	if err != nil || durationStr == "" {
		return defaultExpiration, err
	}

	minutes, err := strconv.Atoi(durationStr)
	if err != nil {
		return defaultExpiration, err
	}

	return time.Duration(minutes) * time.Minute, nil
}

func IsExpired(roomID string) (bool, error) {
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

// Creates a reverse index mapping invitationKey to roomID
func InvitationCodeReverseIndex(invitationKey, roomID string) error {
	defaultExpiration := 30 * time.Minute

	expiresIn, err := getExpirationDuration(roomID)
	if err != nil {
		expiresIn = defaultExpiration
	}

	err = rdb.Client().Set(rdb.Context(), "invitationCode:"+invitationKey, roomID, expiresIn).Err()
	if err != nil {
		return fmt.Errorf("failed to create invkey reverse index %w", err)
	}

	return nil
}

// Checks for any existing room using the invKey reverse index mapped to roomID
func ValidateInvitation(code string) (bool, string, error) {
	roomID, err := rdb.Client().Get(rdb.Context(), "invitationCode:"+code).Result()
	if err != nil {
		if err == redis.Nil {
			return false, "", err
		} else {
			return false, "", err
		}
	}

	return true, roomID, nil
}
