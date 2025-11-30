package api

import (
	"net/http"
	api "server/api/handlers"
	"server/internal/chat"
)

type API struct {
	Chat *chat.Service
}

func (a *API) InitializeRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /create-room", CorsMiddleware(http.HandlerFunc(api.CreateRoomHandler)))
	mux.HandleFunc("/join-room/{room_id}", WithRoomValidation(api.JoinRoomHandler))
	mux.HandleFunc("/start-call/{room_id}", WithRoomValidation(api.StartCallHandler))
	mux.HandleFunc("/leave-call/{room_id}", WithRoomValidation(api.LeaveCallHandler))
	mux.HandleFunc("/assign-host/{room_id}", WithRoomValidation(api.AssignHostHandler))
	mux.HandleFunc("/exit-room/{room_id}", WithRoomValidation(api.ExitRoomHandler))
	mux.HandleFunc("/kill-room/{room_id}", WithRoomValidation(api.KillRoomHandler))

	mux.HandleFunc("/lvk-token/{room_id}", WithRoomValidation(api.LivekitTokenHandler))

	mux.HandleFunc("/set-participant-call-data/{room_id}", WithRoomValidation(api.SetParticipantCallDataHandler))
	mux.HandleFunc("/set-session/{room_id}", WithRoomValidation(api.SetSessionHandler))

	mux.HandleFunc("/room-info/{room_id}", WithRoomValidation(api.GetRoomInfoHandler))
	mux.HandleFunc("/call/{room_id}", WithRoomValidation(api.GetCallStateHandler))
	mux.HandleFunc("/get-me/{room_id}", WithRoomValidation(api.GetMeHandler))
	mux.HandleFunc("/participants/{room_id}", WithRoomValidation(api.GetParticipantsHandler))

	mux.HandleFunc("GET /invitation-code/{room_id}", WithRoomValidation(api.GetInvitationHandler))
	mux.HandleFunc("GET /sse-invitation-update/{room_id}", WithRoomValidation(api.SSEInvitationHandler))

	mux.HandleFunc("/settings/{room_id}", WithRoomValidation(api.GetRoomSettingsHandler))
	mux.HandleFunc("/update-settings/{room_id}", WithRoomValidation(api.UpdateRoomSettingsHandler))

	mux.HandleFunc("/authorization", CorsMiddleware(http.HandlerFunc(api.AuthorizationHandler)))
	mux.HandleFunc("/refresh-token", CorsMiddleware(http.HandlerFunc(api.RefreshTokenHandler)))

	mux.HandleFunc("/ws/chat/{room_id}", WithRoomValidation(func(w http.ResponseWriter, r *http.Request) {
		api.MessageExchange(w, r, a.Chat)
	}))

	mux.HandleFunc("/ws/participants/{room_id}", WithRoomValidation(api.ParticipantsHandler))

	mux.HandleFunc("/ws/settings-broadcast/{room_id}", WithRoomValidation(api.SettingsBroadcastHandler))
	mux.HandleFunc("/ws/user-events/{room_id}", WithRoomValidation(api.UserEventHandler))
	mux.HandleFunc("/ws/system-events/{room_id}", WithRoomValidation(api.SystemEventsHandler))

	return mux
}
