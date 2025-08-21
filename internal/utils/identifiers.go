package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/google/uuid"
)

func GenerateParticipantID() string {
	// We are shortening the user id by encoding it to base64 making data management in redis easier
	id := uuid.New()
	encodedHostID := base64.URLEncoding.EncodeToString(id[:])
	encodedHostID = encodedHostID[:22]

	return encodedHostID
}

func GenerateSessionID() (string, error) {
	const size = 32

	b := make([]byte, size)

	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate session ID: %w", err)
	}

	return base64.URLEncoding.EncodeToString(b), nil
}

// ValidateUUID checks if a string is a valid UUID v4
func ValidateUUID(uuidStr string) bool {
	_, err := uuid.Parse(uuidStr)
	return err == nil
}
