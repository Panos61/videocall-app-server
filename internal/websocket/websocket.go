package websocket

import (
	"sync"

	"github.com/gorilla/websocket"
)

type WSConnection struct {
	Socket *websocket.Conn
	mu     sync.Mutex
}

func (c *WSConnection) Send(message any) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Socket == nil {
		return websocket.ErrCloseSent
	}

	return c.Socket.WriteJSON(message)
}

func (c *WSConnection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Socket != nil {
		err := c.Socket.Close()
		c.Socket = nil
		return err
	}
	return nil
}

type WSConnectionPool struct {
	rooms       map[string]map[string]*WSConnection
	roomLeaders map[string]string
	mu          sync.RWMutex
}

func NewWSConnectionPool() *WSConnectionPool {
	return &WSConnectionPool{
		rooms:       make(map[string]map[string]*WSConnection),
		roomLeaders: make(map[string]string),
	}
}

func (c *WSConnectionPool) AddConnection(roomID, connectionID string, conn *WSConnection) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.rooms[roomID] == nil {
		c.rooms[roomID] = make(map[string]*WSConnection)
	}

	c.rooms[roomID][connectionID] = conn

	if _, hasLeader := c.roomLeaders[roomID]; !hasLeader {
		c.roomLeaders[roomID] = connectionID
	}
}

func (c *WSConnectionPool) RemoveConnection(roomID, connectionID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if room, exists := c.rooms[roomID]; exists {
		delete(room, connectionID)

		// If the leader left, assign new leader
		if c.roomLeaders[roomID] == connectionID {
			delete(c.roomLeaders, roomID)

			// Pick first remaining connection as new leader
			for newLeaderID := range room {
				c.roomLeaders[roomID] = newLeaderID
				break
			}
		}

		if len(room) == 0 {
			delete(c.rooms, roomID)
			delete(c.roomLeaders, roomID)
		}
	}
}

func (c *WSConnectionPool) Broadcast(roomID, senderID string, message any) {
	c.mu.RLock()
	room, exists := c.rooms[roomID]
	if !exists {
		c.mu.RUnlock()
		return
	}

	connections := make(map[string]*WSConnection)
	for connID, conn := range room {
		if connID != senderID {
			connections[connID] = conn
		}
	}
	c.mu.RUnlock()
	for connID, conn := range connections {
		if err := conn.Send(message); err != nil {
			c.RemoveConnection(roomID, connID)
			conn.Close()
		}
	}
}

func (c *WSConnectionPool) BroadcastToAll(roomID string, message any) {
	c.Broadcast(roomID, "", message)
}

func (c *WSConnectionPool) GetConnection(roomID, connectionID string) (*WSConnection, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if room, exists := c.rooms[roomID]; exists {
		if conn, exists := room[connectionID]; exists {
			return conn, true
		}
	}

	return nil, false
}

func (c *WSConnectionPool) SendToConnection(roomID, connectionID string, message any) error {
	if conn, exists := c.GetConnection(roomID, connectionID); exists {
		return conn.Send(message)
	}

	return nil
}

func (c *WSConnectionPool) IsRoomLeader(roomID, connectionID string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	leader, exists := c.roomLeaders[roomID]
	return exists && leader == connectionID
}
