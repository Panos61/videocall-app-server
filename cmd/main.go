package main

import (
	"log"
	"server/internal"
	"server/internal/rdb"
	"server/internal/rmq"
)

func main() {
	// init redis
	rdb.InitRedisClient()

	// init rabbitMQ
	rmqClient, err := rmq.InitRabbitMQ("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal(err)
	}
	defer rmqClient.Close()

	internal.Run()
}
