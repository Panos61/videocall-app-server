package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"server/internal/settings"
	"server/internal/utils"
)

func GetRoomSettingsHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("room_id")

	settings, err := settings.GetRoomSettings(roomID)
	if err != nil {
		http.Error(w, "failed to get room settings", http.StatusInternalServerError)
		return
	}

	utils.JSONResponse(w, settings, http.StatusOK)
}

func UpdateRoomSettingsHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("room_id")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}

	var requestBody struct {
		Settings settings.Settings `json:"settings"`
	}

	err = json.Unmarshal(body, &requestBody)
	if err != nil {
		http.Error(w, "failed to unmarshal request body", http.StatusBadRequest)
		return
	}

	settingsData := requestBody.Settings

	updatedSettings, err := settings.UpdateRoomSettings(roomID, settingsData)
	if err != nil {
		http.Error(w, "failed to update room settings", http.StatusInternalServerError)
		return
	}

	fmt.Println("UPDATED SETTINGS", updatedSettings)

	utils.JSONResponse(w, updatedSettings, http.StatusOK)
}

func SettingsBroadcastHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("room_id")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "failed to upgrade to websocket", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	settings.SettingsSubscription(roomID, conn)
}
