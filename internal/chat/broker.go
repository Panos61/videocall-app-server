package chat

import (
	"server/internal/rmq"

	"github.com/rabbitmq/amqp091-go"
)

const chatExchange = "chat"

type MessageBroker interface {
	PublishToRoom(roomID string, body []byte) error
	SubscribeRoom(roomID string) (<-chan amqp091.Delivery, error)
}

type RMQBroker struct {
	ch *amqp091.Channel
}

func NewRMQBroker(r *rmq.RMQ) (*RMQBroker, error) {
	if err := r.Channel.ExchangeDeclare(chatExchange, "topic", true, false, false, false, nil); err != nil {
		return nil, err
	}

	return &RMQBroker{ch: r.Channel}, nil
}

func (b *RMQBroker) PublishToRoom(roomID string, body []byte) error {
	routingKey := "room." + roomID
	return b.ch.Publish(chatExchange, routingKey, false, false, amqp091.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp091.Persistent, // check this (optional)
		Body:         body,
	})
}

func (b *RMQBroker) SubscribeRoom(roomID string) (<-chan amqp091.Delivery, error) {
	routingKey := "room." + roomID

	q, err := b.ch.QueueDeclare(
		"",
		false,
		true,
		true,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	if err := b.ch.QueueBind(q.Name, routingKey, chatExchange, false, nil); err != nil {
		return nil, err
	}

	msgs, err := b.ch.Consume(q.Name, "", false, true, false, false, nil)
	if err != nil {
		return nil, err
	}

	return msgs, nil
}
