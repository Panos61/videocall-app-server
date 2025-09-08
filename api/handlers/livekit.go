package api

import (
	"encoding/json"
	"io"
	"net/http"

	"server/internal/config"
	"server/internal/utils"

	"github.com/livekit/protocol/auth"
)

func LivekitTokenHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("room_id")
	if roomID == "" {
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var requestBody struct {
		SessionID string `json:"session_id"`
	}

	err = json.Unmarshal(body, &requestBody)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	sessionID := requestBody.SessionID
	if sessionID == "" {
		http.Error(w, "missing session_id", http.StatusBadRequest)
		return
	}

	livekitToken, err := createLivekitToken(roomID, sessionID)
	if err != nil {
		http.Error(w, "failed to create livekit token", http.StatusInternalServerError)
		return
	}

	utils.JSONResponse(w, livekitToken, http.StatusOK)
}

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
