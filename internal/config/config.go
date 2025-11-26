package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Livekit Livekit
	Redis   Redis
	RMQ     RMQ
}
type Livekit struct {
	URL       string
	ApiKey    string
	ApiSecret string
}

type Redis struct {
	URL      string
	Password string
	DB       int
}

type RMQ struct {
	URL string
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	redisURL := os.Getenv("REDIS_URL")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisDB := os.Getenv("REDIS_DB")
	rmqURL := os.Getenv("RMQ_URL")
	livekitURL := os.Getenv("LIVEKIT_URL")
	livekitApiKey := os.Getenv("LIVEKIT_API_KEY")
	livekitApiSecret := os.Getenv("LIVEKIT_API_SECRET")

	requiredVars := map[string]string{
		"REDIS_URL": redisURL,
		// "REDIS_PASSWORD":     redisPassword,
		"REDIS_DB":           redisDB,
		"RMQ_URL":            rmqURL,
		"LIVEKIT_URL":        livekitURL,
		"LIVEKIT_API_KEY":    livekitApiKey,
		"LIVEKIT_API_SECRET": livekitApiSecret,
	}

	var missing []string
	for key, value := range requiredVars {
		if value == "" {
			missing = append(missing, key)
		}
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required environment variables: %v", missing)
	}

	redisDBInt, err := strconv.Atoi(redisDB)
	if err != nil {
		return nil, fmt.Errorf("error converting redis db to int: %w", err)
	}

	config := &Config{
		Redis: Redis{
			URL:      redisURL,
			Password: redisPassword,
			DB:       redisDBInt,
		},
		RMQ: RMQ{
			URL: rmqURL,
		},
		Livekit: Livekit{
			URL:       livekitURL,
			ApiKey:    livekitApiKey,
			ApiSecret: livekitApiSecret,
		},
	}

	return config, nil
}
