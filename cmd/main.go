package main

import (
	"github.com/joho/godotenv"
	"log"
	"mini-broker/internal/app"
	"net/http"
)

func main() {
	envErr := godotenv.Load()
	if envErr != nil {
		log.Fatal("Error loading .env file")
	}
	app.Handler()
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}
