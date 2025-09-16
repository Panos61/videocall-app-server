package internal

import (
	"log"
	"net/http"
	"server/api"

	"github.com/joho/godotenv"
)

func Run() {
	router := api.InitializeRoutes()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file")
	}

	log.Println("Server up and running on port 8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Printf("Failed to start server: %s\n", err)
	}

}
