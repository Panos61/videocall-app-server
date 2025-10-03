package rmq

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

var RMQClient *RMQ

type RMQ struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
}

func InitRabbitMQ(url string) (*RMQ, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		log.Fatalf("Failed to open rabbitmq channel: %v", err)
	}

	log.Println("Connected to RabbitMQ successfully.")
	return &RMQ{
		Conn:    conn,
		Channel: ch,
	}, nil
}

func (r *RMQ) Close() error {
	var err error
	if r.Channel != nil {
		if closeErr := r.Channel.Close(); closeErr != nil {
			err = closeErr
		}
	}

	if r.Conn != nil {
		if closeErr := r.Conn.Close(); closeErr != nil {
			err = closeErr
		}
	}

	return err
}
