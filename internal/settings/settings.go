package settings

import (
	"fmt"
	"server/internal/rdb"
	"server/internal/room"
)

type Settings struct {
	InvitationExpiry string `json:"invitation_expiry"`
	InvitePermission bool   `json:"invite_permission"`
}

func GetRoomSettings(roomID string) (Settings, error) {
	invitationExpiry, err := room.GetExpiry(roomID)
	if err != nil {
		return Settings{}, err
	}

	settings := Settings{
		InvitationExpiry: invitationExpiry,
		InvitePermission: false,
	}

	fmt.Println("SETTINGS CONTROLLER", settings)

	return settings, nil
}

func UpdateRoomSettings(roomID string, settings Settings) (Settings, error) {
	_, err := rdb.Client().HSet(rdb.Context(), "room:"+roomID+":settings", map[string]any{
		"invitation_expiry": settings.InvitationExpiry,
		"invite_permission": settings.InvitePermission,
	}).Result()

	if err != nil {
		return Settings{}, fmt.Errorf("error updating room settings: %w", err)
	}

	return settings, nil
}
