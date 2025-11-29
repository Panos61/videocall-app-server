package api

import (
	"net/http"
	"server/internal/room"
	"server/internal/utils"
)

func CorsMiddleware(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// log.Printf("CORS Middleware handling request: %s %s", r.Method, r.URL.Path)

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, Pragma, Expires, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			// log.Printf("Preflight request handled: %s %s", r.Method, r.URL.Path)
			return
		}

		next.ServeHTTP(w, r)
	}
}

// ValidateRoomIDMiddleware validates that the room_id path parameter is a valid UUID
func ValidateRoomIDMiddleware(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roomID := r.PathValue("room_id")

		// Only validate if room_id is present in the path
		if roomID != "" && !utils.ValidateUUID(roomID) {
			http.Error(w, "invalid room ID format", http.StatusBadRequest)
			return
		}

		roomExists, _ := room.GetRoomID(roomID)
		if roomExists == "" {
			http.Error(w, "room not found", http.StatusNotFound)
			return
		}

		next.ServeHTTP(w, r)
	}
}

// RoomMiddleware applies both CORS and room ID validation middleware
func WithRoomValidation(handler http.HandlerFunc) http.HandlerFunc {
	return CorsMiddleware(ValidateRoomIDMiddleware(http.HandlerFunc(handler)))
}
