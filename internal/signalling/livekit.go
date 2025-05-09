package signalling

import (
	"server/internal/config"

	"github.com/livekit/protocol/auth"
)

func createLivekitToken(roomID, sessionID string) (string, error) {
	config, err := config.LoadConfig()
	if err != nil {
		return "", err
	}

	LIVEKIT_API_KEY := config.Livekit.ApiKey
	LIVEKIT_API_SECRET := config.Livekit.ApiSecret

	token := auth.NewAccessToken(LIVEKIT_API_KEY, LIVEKIT_API_SECRET)
	token.SetIdentity(sessionID)
	token.SetVideoGrant(&auth.VideoGrant{
		RoomJoin: true,
		Room:     roomID,
	})

	return token.ToJWT()
}
