package userevents

import (
	"server/internal/events"
	"server/internal/participant"
)

type MediaControlEvent struct {
	AudioEnabled bool `json:"audio"`
	VideoEnabled bool `json:"video"`
}

type SyncMediaStateEvent map[string]MediaControlEvent

type ShareScreenEvent struct {
	Username string `json:"username"`
	TrackSID string `json:"track_sid"`
	Active   bool   `json:"active"`
}

type RaisedHandEvent struct {
	HandRaised bool   `json:"raised_hand"`
	Username   string `json:"username"`
}

type ReactionEvent struct {
	ReactionType string `json:"reaction_type"`
	Username     string `json:"username"`
}

func handleReactionSent(roomID, participantID string, reactionData ReactionEvent) {
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

func handleRaisedHandSent(roomID, participantID string, raisedHandData RaisedHandEvent) {
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

func handleShareScreenStarted(roomID, participantID string, shareScreenData ShareScreenEvent) {
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
