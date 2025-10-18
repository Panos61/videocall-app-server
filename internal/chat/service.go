package chat

import (
	"encoding/json"

	"github.com/rabbitmq/amqp091-go"
)

type Service struct {
	broker MessageBroker
}

func NewService(b MessageBroker) *Service {
	return &Service{broker: b}
}

func (s *Service) Publish(roomID string, out OutboundMsg) error {
	body, err := json.Marshal(out)
	if err != nil {
		return err
	}

	return s.broker.PublishToRoom(roomID, body)
}

func (s *Service) Subscribe(roomID string) (delivery <-chan amqp091.Delivery, err error) {
	return s.broker.SubscribeRoom(roomID)
}
