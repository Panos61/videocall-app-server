package api

import (
	"encoding/json"
	"log"
	"net/http"
	"server/internal/chat"

	"github.com/gorilla/websocket"
)

func MessageExchange(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	// Create channels for coordination
	// done := make(chan struct{})

	// Handle incoming messages from WebSocket (send to RabbitMQ)
	go func() {
		// defer close(done)
		for {
			var message chat.Message
			err := conn.ReadJSON(&message)
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket error: %v", err)
				}
				return
			}

			if err := chat.GlobalService.SendMessage(message); err != nil {
				log.Printf("Error sending message: %v", err)
			}
		}
	}()

	// Handle outgoing messages (consume from RabbitMQ and send to WebSocket)
	messages, err := chat.GlobalService.ReceiveMessages()
	if err != nil {
		log.Printf("Error setting up message consumer: %v", err)
		return
	}

	// Forward messages from RabbitMQ to WebSocket
	go func() {
		for msgBytes := range messages {
			var message chat.Message
			if err := json.Unmarshal(msgBytes, &message); err != nil {
				log.Printf("Error unmarshaling message: %v", err)
				continue
			}

			if err := conn.WriteJSON(message); err != nil {
				log.Printf("Error writing to WebSocket: %v", err)
				return
			}
			// select {
			// // case <-done:
			// // 	return // Connection closed, stop sending
			// default:
			// 	if err := conn.WriteJSON(message); err != nil {
			// 		log.Printf("Error writing to WebSocket: %v", err)
			// 		return
			// 	}
			// }
		}
	}()
}
