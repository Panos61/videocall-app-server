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
	fmt.Println("SETTINGS", settings)
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

	fmt.Println("BODY", string(body))

	var settingsData settings.Settings

	err = json.Unmarshal(body, &settingsData)
	if err != nil {
		http.Error(w, "failed to unmarshal request body", http.StatusBadRequest)
		return
	}

	updatedSettings, err := settings.UpdateRoomSettings(roomID, settingsData)
	if err != nil {
		http.Error(w, "failed to update room settings", http.StatusInternalServerError)
		return
	}

	utils.JSONResponse(w, updatedSettings, http.StatusOK)
}
