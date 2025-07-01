package api

import (
	"net/http"
	"server/internal/settings"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Connection struct {
	Socket *websocket.Conn
	mu     sync.Mutex
}

type AuthMessage struct {
	Type      string `json:"type"`
	Token     string `json:"token"`
	SessionID string `json:"sessionID"`
}

type MediaState struct {
	Audio bool `json:"audio"`
	Video bool `json:"video"`
}

type SettingsMessage struct {
	Settings settings.Settings `json:"settings"`
}

type Message struct {
	Type        string            `json:"type"`
	RoomID      string            `json:"roomID"`
	SessionID   string            `json:"sessionID"`
	Token       string            `json:"token"`
	Description string            `json:"description"`
	To          string            `json:"to"`
	Media       MediaState        `json:"media"`
	Settings    settings.Settings `json:"settings"`
}

var connectionPool = make(map[string]map[string]*Connection)
var connectionsMutex sync.Mutex

func (c *Connection) Send(message Message) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Socket.WriteJSON(message)
}
