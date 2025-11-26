package main

import (
	"log"
	"server/internal"
	"server/internal/config"
	"server/internal/rdb"
	"server/internal/rmq"
)

func main() {
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	// init redis
	rdb.InitRedisClient()

	// init rabbitMQ
	rmqClient, err := rmq.InitRabbitMQ(config.RMQ.URL)
	if err != nil {
		log.Fatal(err)
	}
	defer rmqClient.Close()

	internal.Run(rmqClient)
}
