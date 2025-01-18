package api

import (
	"fmt"
	"net/http"
	"server/internal/room"
	"sync"
	"time"
)

var roomClients = make(map[string][]http.ResponseWriter)
var roomClientsLock = sync.Mutex{}

func addClient(roomID string, w http.ResponseWriter) {
	roomClientsLock.Lock()
	defer roomClientsLock.Unlock()

	roomClients[roomID] = append(roomClients[roomID], w)
}

func removeClient(roomID string, w http.ResponseWriter) {
	roomClientsLock.Lock()
	defer roomClientsLock.Unlock()

	clients := roomClients[roomID]
	for i, client := range clients {
		if client == w {
			roomClients[roomID] = append(clients[:i], clients[i+1:]...)
			break
		}
	}

	// Clean up room if no clients are left
	if len(roomClients[roomID]) == 0 {
		delete(roomClients, roomID)
	}
}

func broadcastToClients(roomID, message string) {
	roomClientsLock.Lock()
	defer roomClientsLock.Unlock()

	clients := roomClients[roomID]
	for _, client := range clients {
		_, err := fmt.Fprintf(client, "event: update\ndata: %s\n\n", message)
		if err != nil {
			// Remove client on error
			removeClient(roomID, client)
		} else {
			client.(http.Flusher).Flush()
		}
	}
}

func SSEInvitationHandler(w http.ResponseWriter, r *http.Request) {
	_, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	roomID := r.PathValue("room_id")
	if roomID == "" {
		http.Error(w, "room id is required", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	addClient(roomID, w)
	defer removeClient(roomID, w)

	// Listen for client disconnect
	notify := r.Context().Done()

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-notify:
			return
		case <-ticker.C:
			isExpired, err := room.IsExpired(roomID)
			if err != nil {
				broadcastToClients(roomID, fmt.Sprintf("event: error\ndata: %v\n\n", err))
				return
			}

			if isExpired {
				newCode := room.GenerateInvitationCode(roomID)
				invitationURL, err := room.SetInvitation(roomID, newCode)
				if err != nil {
					broadcastToClients(roomID, fmt.Sprintf("event: error\ndata: %v\n\n", err))
					return
				}

				broadcastToClients(roomID, invitationURL)

				err = room.InvitationCodeReverseIndex(newCode, roomID)
				if err != nil {
					broadcastToClients(roomID, fmt.Sprintf("event: error\ndata: %v\n\n", err))
					return
				}
			}
		}
	}
}
