package main

import (
	"bookstore/internal/db"
	"log"
	"net/http"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	client, mongoDB, err := db.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(db.Bg())

	mux := http.NewServeMux()
	RegisterRoutes(mux, mongoDB)

	log.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
