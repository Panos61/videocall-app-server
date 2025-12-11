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
		ParticipantID string `json:"participant_id"`
	}

	err = json.Unmarshal(body, &requestBody)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	participantID := requestBody.ParticipantID
	if participantID == "" {
		http.Error(w, "missing participant_id", http.StatusBadRequest)
		return
	}

	livekitToken, err := createLivekitToken(roomID, participantID)
	if err != nil {
		http.Error(w, "failed to create livekit token", http.StatusInternalServerError)
		return
	}

	utils.JSONResponse(w, livekitToken, http.StatusOK)
}

func createLivekitToken(roomID, participantID string) (string, error) {
	config, err := config.LoadConfig()
	if err != nil {
		return "", err
	}

	LIVEKIT_API_KEY := config.Livekit.ApiKey
	LIVEKIT_API_SECRET := config.Livekit.ApiSecret

	token := auth.NewAccessToken(LIVEKIT_API_KEY, LIVEKIT_API_SECRET)
	token.SetIdentity(participantID)
	token.SetVideoGrant(&auth.VideoGrant{
		RoomJoin: true,
		Room:     roomID,
	})

	return token.ToJWT()
}
