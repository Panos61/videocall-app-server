package config

import (
	"fmt"
	"os"
)

type Config struct {
	Livekit Livekit
	RMQ     RMQ
}
type Livekit struct {
	URL       string
	ApiKey    string
	ApiSecret string
}

type RMQ struct {
	URL string
}

func LoadConfig() (*Config, error) {
	url := os.Getenv("LIVEKIT_URL")
	apiKey := os.Getenv("LIVEKIT_API_KEY")
	apiSecret := os.Getenv("LIVEKIT_API_SECRET")

	rmqURL := os.Getenv("RMQ_URL")

	if url == "" || apiKey == "" || apiSecret == "" || rmqURL == "" {
		return nil, fmt.Errorf("missing required environment variables")
	}

	config := &Config{
		Livekit: Livekit{
			URL:       url,
			ApiKey:    apiKey,
			ApiSecret: apiSecret,
		},
		RMQ: RMQ{
			URL: rmqURL,
		},
	}

	return config, nil
}
