package utils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	ParticipantID string `json:"participant_id"`
	IsHost        bool   `json:"is_host"`
	jwt.RegisteredClaims
}

var secretKey = []byte("secret_key")

func GenerateJWT(participantID string, isHost bool) (string, error) {
	expirationTime := time.Now().Add(30 * time.Minute)

	claims := &Claims{
		ParticipantID: participantID,
		IsHost:        isHost,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(secretKey)

	if err != nil {
		return "", err
	}

	return tokenStr, nil
}

func ValidateToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil {
		if valErr, ok := err.(*jwt.ValidationError); ok {
			if valErr.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
				return claims, fmt.Errorf("token is expired")
			}
		}
		return nil, fmt.Errorf("invalid token")
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
