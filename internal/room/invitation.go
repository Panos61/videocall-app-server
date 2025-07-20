package room

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"regexp"
	"server/internal/rdb"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var codeRegex = regexp.MustCompile("^[a-zA-Z0-9_-]{16}$")

func GenerateCode() string {
	buffer := make([]byte, 12)

	_, err := rand.Read(buffer)
	if err != nil {
		log.Printf("Error generating invitation code: %s\n", err)
		return ""
	}

	return base64.URLEncoding.EncodeToString(buffer)
}

func UpdateInvitationCode(roomID string) (string, error) {
	code := GenerateCode()

	durationStr, err := GetInvitationExpiry(roomID)
	if err != nil {
		return "", err
	}

	duration, err := strconv.Atoi(durationStr)
	if err != nil {
		return "", err
	}

	ttl := time.Duration(duration) * time.Minute
	err = rdb.Client().Set(rdb.Context(), "room:"+roomID+":invitation", code, ttl).Err()
	if err != nil {
		return "", err
	}

	return code, nil
}

func GetInvitationCode(roomID string) (string, error) {
	invitationCode, err := rdb.Client().Get(rdb.Context(), "room:"+roomID+":invitation").Result()
	if err != nil {
		return "", err
	}

	return invitationCode, nil
}

func GetInvitationExpiry(roomID string) (string, error) {
	duration, err := rdb.Client().HGet(rdb.Context(), "room:"+roomID+":settings", "invitation_expiry").Result()
	if err != nil {
		return "30", err
	}

	return duration, nil
}

func HasExpired(roomID, expectedCode string) (bool, error) {
	actualCode, err := rdb.Client().Get(rdb.Context(), "room:"+roomID+":invitation").Result()
	if err != nil {
		if err == redis.Nil {
			return true, nil
		}
		return false, err
	}

	return actualCode != expectedCode, nil
}
func matchesFormat(invitationCode string) bool {
	return codeRegex.MatchString(invitationCode)
}

func IsCodeValid(roomID, invitationCode string) (bool, bool, error) {
	hasExpired, err := HasExpired(roomID, invitationCode)
	if err != nil || hasExpired {
		return true, false, err
	}

	matchesFormat := matchesFormat(invitationCode)
	if !matchesFormat {
		return false, false, nil
	}

	return false, true, nil
}
