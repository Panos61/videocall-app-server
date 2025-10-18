package internal

import (
	"log"
	"net/http"
	"server/api"
	"server/internal/chat"
	"server/internal/rmq"

	"github.com/joho/godotenv"
)

func Run(rmqClient *rmq.RMQ) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file")
	}

	broker, err := chat.NewRMQBroker(rmqClient)
	if err != nil {
		log.Fatalf("chat broker error: %v", err)
	}
	svc := chat.NewService(broker)

	a := &api.API{Chat: svc}
	mux := a.InitializeRoutes()

	addr := ":8080"
	log.Printf("HTTP listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
