package api

import (
	"net/http"
	"server/internal/room"
	"server/internal/utils"
)

func CorsMiddleware(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		allowed := map[string]bool{
			"https://app.panos-dev.com": true,
			"http://localhost:5173":     true,
		}

		if allowed[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		w.Header().Set("Access-Control-Allow-Headers",
			"Content-Type, Authorization, X-Requested-With, Accept, Origin, Cache-Control, Pragma")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
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
