package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"server/internal/chat"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

func MessageExchange(w http.ResponseWriter, r *http.Request, chatSvc *chat.Service) {
	roomID := r.PathValue("room_id")
	if roomID == "" {
		http.Error(w, "room not found", http.StatusBadRequest)
		return
	}

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	deliveries, err := chatSvc.Subscribe(roomID)
	if err != nil {
		http.Error(w, "rmq subscription failed", http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// client -> publish
	go func() {
		for {
			var inMsg chat.InboundMsg
			if err := conn.ReadJSON(&inMsg); err != nil {
				cancel()
				return
			}

			fmt.Println("inMsg", inMsg)
			outMsg := chat.OutboundMsg{
				ID:        uuid.NewString(),
				Payload:   inMsg.Payload,
				User:      inMsg.User,
				Timestamp: time.Now().UnixMilli(),
			}

			if err := chatSvc.Publish(roomID, outMsg); err != nil {
				cancel()
				return
			}
		}
	}()

	// deliveries -> client
	for {
		select {
		case <-ctx.Done():
			return
		case d, ok := <-deliveries:
			if !ok {
				return
			}

			var outMsg chat.OutboundMsg
			if err := json.Unmarshal(d.Body, &outMsg); err != nil {
				_ = d.Nack(false, false)
			}

			if err := conn.WriteJSON(outMsg); err != nil {
				_ = d.Nack(false, true) // requeue on ws error
				return
			}
			fmt.Println("outMsg publish", outMsg)
			_ = d.Ack(false)
		}
	}
}
