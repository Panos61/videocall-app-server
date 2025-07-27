package api

import (
	"net/http"
	api "server/api/handlers"
)

func InitializeRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /create-room", CorsMiddleware(http.HandlerFunc(api.CreateRoomHandler)))
	mux.HandleFunc("/join-room/{room_id}", CorsMiddleware(http.HandlerFunc(api.JoinRoomHandler)))
	mux.HandleFunc("/exit-room/{room_id}", CorsMiddleware(http.HandlerFunc(api.ExitRoomHandler)))
	mux.HandleFunc("/start-call/{room_id}", CorsMiddleware(http.HandlerFunc(api.StartCallHandler)))
	mux.HandleFunc("/leave-call/{room_id}", CorsMiddleware(http.HandlerFunc(api.LeaveCallHandler)))

	mux.HandleFunc("/call/{room_id}", CorsMiddleware(http.HandlerFunc(api.GetCallStateHandler)))
	mux.HandleFunc("/set-participant-call-data/{room_id}", CorsMiddleware(http.HandlerFunc(api.SetParticipantCallDataHandler)))
	mux.HandleFunc("/set-session/{room_id}", CorsMiddleware(http.HandlerFunc(api.SetSessionHandler)))

	mux.HandleFunc("/room-info/{room_id}", CorsMiddleware(http.HandlerFunc(api.GetRoomInfoHandler)))
	mux.HandleFunc("/get-me/{room_id}", CorsMiddleware(http.HandlerFunc(api.GetMeHandler)))
	mux.HandleFunc("/participants/{room_id}", CorsMiddleware(http.HandlerFunc(api.GetParticipantsHandler)))

	mux.HandleFunc("GET /invitation-code/{room_id}", CorsMiddleware(http.HandlerFunc(api.GetInvitationHandler)))
	mux.HandleFunc("GET /sse-invitation-update/{room_id}", CorsMiddleware(http.HandlerFunc(api.SSEInvitationHandler)))

	mux.HandleFunc("/settings/{room_id}", CorsMiddleware(http.HandlerFunc(api.GetRoomSettingsHandler)))
	mux.HandleFunc("/update-settings/{room_id}", CorsMiddleware(http.HandlerFunc(api.UpdateRoomSettingsHandler)))

	mux.HandleFunc("/authorization", CorsMiddleware(http.HandlerFunc(api.AuthorizationHandler)))
	mux.HandleFunc("/refresh-token", CorsMiddleware(http.HandlerFunc(api.RefreshTokenHandler)))

	mux.HandleFunc("/ws/signalling/{room_id}", CorsMiddleware(http.HandlerFunc(api.SessionHandler)))
	mux.HandleFunc("/ws/call/{room_id}", CorsMiddleware(http.HandlerFunc(api.CallBroadcastHandler)))
	mux.HandleFunc("/ws/media/{room_id}", CorsMiddleware(http.HandlerFunc(api.MediaHandler)))
	mux.HandleFunc("/ws/settings-broadcast/{room_id}", CorsMiddleware(http.HandlerFunc(api.SettingsBroadcastHandler)))
	mux.HandleFunc("/ws/guests/{room_id}", CorsMiddleware(http.HandlerFunc(api.GuestsHandler)))
	mux.HandleFunc("/ws/user-events/{room_id}", CorsMiddleware(http.HandlerFunc(api.UserEventNotifyHandler)))

	return mux
}
