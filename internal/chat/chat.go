package chat

import (
	"encoding/json"
)

type Message struct {
	ID        string `json:"message_id"`
	Payload   string `json:"payload"`
	Timestamp int64  `json:"timestamp"`
}

type MessageBroker interface {
	PublishMessage(msg []byte, queue string) error
	ConsumeMessages(queue string) (<-chan []byte, error)
}

type Service struct {
	broker MessageBroker
}

// Global service instance
var GlobalService *Service

func NewService(broker MessageBroker) *Service {
	return &Service{broker: broker}
}

// Initialize global service
func InitGlobalService(broker MessageBroker) {
	GlobalService = NewService(broker)
}

func (s *Service) SendMessage(msg Message) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return s.broker.PublishMessage(body, "chat")
}

func (s *Service) ReceiveMessages() (<-chan []byte, error) {
	return s.broker.ConsumeMessages("chat")
}
