package userevents

import (
	"server/internal/events"
	"server/internal/participant"
)

type MediaState struct {
	AudioEnabled bool `json:"audio"`
	VideoEnabled bool `json:"video"`
}

type SyncMediaStateEvent map[string]MediaState

type ShareScreen struct {
	Username string `json:"username"`
	TrackSID string `json:"track_sid"`
	Active   bool   `json:"active"`
}

type RaisedHand struct {
	HandRaised bool   `json:"raised_hand"`
	Username   string `json:"username"`
}

type Reaction struct {
	ReactionType string `json:"reaction_type"`
	Username     string `json:"username"`
}

func handleMediaState(roomID, participantID string, mediaState MediaState) {
	PublishUserEvent(roomID, UserEvent{
		Type: events.MediaStateUpdated,
		Payload: map[string]any{
			"audio": mediaState.AudioEnabled,
			"video": mediaState.VideoEnabled,
		},
		SessionID: participantID,
	})
}

func handleSyncMedia(roomID, participantID string, syncMediaData SyncMediaStateEvent) {
	PublishUserEvent(roomID, UserEvent{
		Type: events.SyncMedia,
		Payload: map[string]any{
			"media_state": syncMediaData,
		},
		SessionID: participantID,
	})
}

func handleReactionSent(roomID, participantID string, reactionData Reaction) {
	participantData, err := participant.GetParticipant(roomID, participantID)
	if err != nil {
		return
	}

	PublishUserEvent(roomID, UserEvent{
		Type: events.ReactionSent,
		Payload: map[string]any{
			"reaction_type": reactionData.ReactionType,
			"username":      participantData.Username,
		},
	})
}

func handleRaisedHandSent(roomID, participantID string, raisedHandData RaisedHand) {
	participantData, err := participant.GetParticipant(roomID, participantID)
	if err != nil {
		return
	}

	PublishUserEvent(roomID, UserEvent{
		Type: events.RaisedHand,
		Payload: map[string]any{
			"raised_hand": raisedHandData.HandRaised,
			"username":    participantData.Username,
		},
	})
}

func handleShareScreenStarted(roomID, participantID string, shareScreenData ShareScreen) {
	participantData, err := participant.GetParticipant(roomID, participantID)
	if err != nil {
		return
	}

	PublishUserEvent(roomID, UserEvent{
		Type: events.ShareScreenStarted,
		Payload: map[string]any{
			"username":  participantData.Username,
			"track_sid": shareScreenData.TrackSID,
			"active":    shareScreenData.Active,
		},
	})
}
