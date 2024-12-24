package internal

import (
	"log"
	"net/http"
	"server/api"
	"server/internal/turnserver"
)

func Run() {
	router := api.InitializeRoutes()

	go func() {
		log.Println("Starting TURN server on port 3478")
		turnserver.StartTurnServer("0.0.0.0", 3478, "localtest.com")
	}()

	log.Println("Server up and running on port 8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Printf("Failed to start server: %s\n", err)
	}

}
