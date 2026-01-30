package main

import (
	"log"
	"net/http"
	"os"

	"bookstore/internal/db"
	"bookstore/internal/handlers"
	"bookstore/internal/repository"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	client, mongoDB, err := db.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(db.Bg())

	var bookRepo repository.BookRepository = repository.NewBookRepo(mongoDB)
	bookHandler := handlers.NewBookHandler(bookRepo)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	mux.HandleFunc("/books", bookHandler.Books)
	mux.HandleFunc("/books/", bookHandler.BookByID)

	addr := ":8080"
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	}

	log.Println("Server running on", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
