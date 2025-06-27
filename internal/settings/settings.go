package settings

import (
	"fmt"
	"server/internal/rdb"
)

type Settings struct {
	InvitationExpiry string `json:"invitation_expiry"`
	InvitePermission bool   `json:"invite_permission"`
}

func GetRoomSettings(roomID string) (Settings, error) {
	settingsData, err := rdb.Client().HGetAll(rdb.Context(), "room:"+roomID+":settings").Result()
	if err != nil {
		return Settings{}, err
	}

	// fmt.Println("SETTINGS DATA", settingsData)

	settings := Settings{
		InvitationExpiry: settingsData["invitation_expiry"],
		InvitePermission: settingsData["invite_permission"] == "true",
	}

	// fmt.Println("SETTINGS CONTROLLER", settings)

	return settings, nil
}

func UpdateRoomSettings(roomID string, settings Settings) (Settings, error) {
	_, err := rdb.Client().HSet(rdb.Context(), "room:"+roomID+":settings", map[string]any{
		"invitation_expiry": settings.InvitationExpiry,
		"invite_permission": settings.InvitePermission,
	}).Result()
	// fmt.Println("UPDATED SETTINGS", settingsData)

	if err != nil {
		return Settings{}, fmt.Errorf("error updating room settings: %w", err)
	}

	return settings, nil
}
