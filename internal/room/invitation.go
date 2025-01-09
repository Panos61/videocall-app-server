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

func SetInvitationExpiration(roomID string, inviteExpiration string) (bool, error) {
	expiryToStr, err := strconv.Atoi(inviteExpiration)
	if err != nil {
		return false, err
	}

	expiresIn := time.Duration(expiryToStr) * time.Minute

	err = rdb.Client().HSet(rdb.Context(), "room:"+roomID, "expiresIn", time.Now().Add(expiresIn).Format(time.RFC3339)).Err()
	fmt.Println(err)
	if err != nil {
		return false, err
	}

	return true, nil
}

// Sets the invitation URL and expiration time for the room
func SetInvitation(roomID, invitationCode string) (string, error) {
	expirationTime := 20 * time.Second
	invitationURL := BuildInvitationURL(roomID, invitationCode)

	_, err := rdb.Client().HSet(rdb.Context(), "room:"+roomID, map[string]interface{}{
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

func GetCurrentInvitation(roomID string) (string, error) {
	invitation, err := rdb.Client().HGet(rdb.Context(), roomID, "invitation").Result()

	if invitation == "" || err != nil {
		return "", err
	}

	return invitation, nil
}

// Creates a reverse index mapping invitation key to roomID
func CreateInvitationIndex(invitation, roomID string) error {
	expires := 20 * time.Second

	err := rdb.Client().Set(rdb.Context(), "invitation:"+invitation, roomID, expires).Err()
	if err != nil {
		return fmt.Errorf("failed to create invkey reverse index %w", err)
	}

	return nil
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
