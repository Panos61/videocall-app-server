package api

import (
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// return true
		o := r.Header.Get("Origin")
		return o == "https://app.panos-dev.com" || o == "http://localhost:5173"
	},
}

type Connection struct {
	Socket *websocket.Conn
}
