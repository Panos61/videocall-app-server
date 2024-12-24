package api

import (
	"net/http"
	api "server/api/handlers"
	"server/internal/signalling"
)

func InitializeRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /create-room", CorsMiddleware(http.HandlerFunc(api.CreateRoomHandler)))
	mux.HandleFunc("/join-room", CorsMiddleware(http.HandlerFunc(api.JoinRoomHandler)))
	mux.HandleFunc("/leave-room/{id}", CorsMiddleware(http.HandlerFunc(api.LeaveRoomHandler)))

	mux.HandleFunc("/start-call/{room_id}", CorsMiddleware(http.HandlerFunc(api.StartCallHandler)))
	mux.HandleFunc("/set-session/{room_id}", CorsMiddleware(http.HandlerFunc(api.SetSessionHandler)))

	mux.HandleFunc("/update-user-media/{room_id}", CorsMiddleware(http.HandlerFunc(api.UpdateUserMediaHandler)))

	mux.HandleFunc("GET /call-participants/{room_id}", CorsMiddleware(http.HandlerFunc(api.GetCallParticipantsHandler)))

	mux.HandleFunc("GET /room-invitation/{id}", CorsMiddleware(http.HandlerFunc(api.SetInvKeyHandler)))
	mux.HandleFunc("GET /sse-key-update/{id}", CorsMiddleware(http.HandlerFunc(api.SSEKeyUpdateHandler)))

	mux.HandleFunc("/ws/signalling/{room_id}", CorsMiddleware(http.HandlerFunc(signalling.SignallingHandler)))

	return mux
}
