package api

import (
	"net/http"
	api "server/api/handlers"
	"server/internal/session"
)

func InitializeRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /create-room", CorsMiddleware(http.HandlerFunc(api.CreateRoomHandler)))
	mux.HandleFunc("/join-room/{room_id}", CorsMiddleware(http.HandlerFunc(api.JoinRoomHandler)))
	mux.HandleFunc("/leave-room/{room_id}", CorsMiddleware(http.HandlerFunc(api.LeaveRoomHandler)))

	mux.HandleFunc("/start-call/{room_id}", CorsMiddleware(http.HandlerFunc(api.StartCallHandler)))
	mux.HandleFunc("/set-session/{room_id}", CorsMiddleware(http.HandlerFunc(api.SetSessionHandler)))
	mux.HandleFunc("/update-settings/{room_id}", CorsMiddleware(http.HandlerFunc(api.InvitationSettingsHandler)))

	mux.HandleFunc("/get-me/{room_id}", CorsMiddleware(http.HandlerFunc(api.GetMeHandler)))
	mux.HandleFunc("GET /call-participants/{room_id}", CorsMiddleware(http.HandlerFunc(api.GetCallParticipantsHandler)))

	mux.HandleFunc("GET /room-invitation/{room_id}", CorsMiddleware(http.HandlerFunc(api.SetInvitationHandler)))
	mux.HandleFunc("GET /sse-invitation-update/{room_id}", CorsMiddleware(http.HandlerFunc(api.SSEInvitationHandler)))
	mux.HandleFunc("/settings/{room_id}", CorsMiddleware(http.HandlerFunc(api.GetSettings)))

	mux.HandleFunc("/validate-invitation", CorsMiddleware(http.HandlerFunc(api.ValidateInvitationHandler)))
	mux.HandleFunc("/refresh-token", CorsMiddleware(http.HandlerFunc(api.RefreshTokenHandler)))

	mux.HandleFunc("/ws/signalling/{room_id}", CorsMiddleware(http.HandlerFunc(session.SessionHandler)))
	// mux.HandleFunc("/ws/signalling/{room_id}", CorsMiddleware(http.HandlerFunc(signalling.SignallingHandler)))
	mux.HandleFunc("/ws/media/{room_id}", CorsMiddleware(http.HandlerFunc(api.MediaHandler)))
	mux.HandleFunc("/ws/user-events/{room_id}", CorsMiddleware(http.HandlerFunc(api.UserEventNotifyHandler)))

	return mux
}
