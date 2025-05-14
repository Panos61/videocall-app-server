package config

import (
	"fmt"
	"os"
)

type Config struct {
	Livekit Livekit
}

type Livekit struct {
	URL       string
	ApiKey    string
	ApiSecret string
}

func LoadConfig() (*Config, error) {
	url := os.Getenv("LIVEKIT_URL")
	apiKey := os.Getenv("LIVEKIT_API_KEY")
	apiSecret := os.Getenv("LIVEKIT_API_SECRET")

	if url == "" || apiKey == "" || apiSecret == "" {
		return nil, fmt.Errorf("missing required environment variables")
	}

	config := &Config{
		Livekit: Livekit{
			URL:       url,
			ApiKey:    apiKey,
			ApiSecret: apiSecret,
		},
	}

	return config, nil
}
