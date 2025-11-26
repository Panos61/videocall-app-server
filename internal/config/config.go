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
	rmqURL := os.Getenv("RMQ_URL")
	livekitURL := os.Getenv("LIVEKIT_URL")
	livekitApiKey := os.Getenv("LIVEKIT_API_KEY")
	livekitApiSecret := os.Getenv("LIVEKIT_API_SECRET")

	if livekitURL == "" || livekitApiKey == "" || livekitApiSecret == "" || rmqURL == "" {
		return nil, fmt.Errorf("missing required environment variables")
	}

	config := &Config{
		Livekit: Livekit{
			URL:       livekitURL,
			ApiKey:    livekitApiKey,
			ApiSecret: livekitApiSecret,
		},
		RMQ: RMQ{
			URL: rmqURL,
		},
	}

	return config, nil
}
