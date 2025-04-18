package rabbitmq

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func InitRabbitMQ() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	log.Println("Connected to RabbitMQ successfully.")
}
