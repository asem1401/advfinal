package main

import (
	"log"
	"net/http"
	"os"

	"bookstore/internal/db"
	"bookstore/internal/handlers"
	"bookstore/internal/logic"
	"bookstore/internal/repository"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	client, mongoDB, err := db.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(db.Bg())

	bookRepo := repository.NewBookRepo(mongoDB)
	bookService := logic.NewBookService(bookRepo)
	bookHandler := handlers.NewBookHandler(bookService)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	mux.HandleFunc("GET /books", bookHandler.Books)
	mux.HandleFunc("POST /books", bookHandler.Books)
	
	mux.HandleFunc("GET /books/{id}", bookHandler.BookByID)
	mux.HandleFunc("PUT /books/{id}", bookHandler.BookByID)
	mux.HandleFunc("DELETE /books/{id}", bookHandler.BookByID)

	addr := ":8080"
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	}

	log.Println("Server running on", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
