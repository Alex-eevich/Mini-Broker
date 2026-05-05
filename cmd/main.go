package main

import (
	"log"
	"mini-broker/internal/app"
	"net/http"
)

func main() {

	app.Handler()
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}
