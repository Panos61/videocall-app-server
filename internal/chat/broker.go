package chat

import (
	"log"
	"server/internal/rmq"
	"sync"

	"github.com/rabbitmq/amqp091-go"
)

type RMQBroker struct {
	rmq      *rmq.RMQ
	consumer <-chan amqp091.Delivery
	once     sync.Once
}

func NewRMQBroker(rmq *rmq.RMQ) *RMQBroker {
	return &RMQBroker{rmq: rmq}
}

func (r *RMQBroker) PublishMessage(msg []byte, queue string) error {
	_, err := r.rmq.Channel.QueueDeclare(queue, false, false, false, false, nil)
	if err != nil {
		return err
	}

	err = r.rmq.Channel.Publish(
		"",
		queue,
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        msg,
		})

	return err
}

func (r *RMQBroker) ConsumeMessages(queue string) (<-chan []byte, error) {
	var err error

	r.once.Do(func() {
		_, declareErr := r.rmq.Channel.QueueDeclare(queue, false, false, false, false, nil)
		if declareErr != nil {
			err = declareErr
			return
		}

		r.consumer, err = r.rmq.Channel.Consume(
			queue,
			"",
			true, // todo: consumer should return ack (autoAck -> false)
			false,
			false,
			false,
			nil,
		)
	})

	if err != nil {
		return nil, err
	}

	output := make(chan []byte, 10)
	go func() {
		defer close(output)
		for msg := range r.consumer {
			select {
			case output <- msg.Body:
			default:
				log.Println("consumer channel is full")
			}
		}
	}()

	return output, nil
}
