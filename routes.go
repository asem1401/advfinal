package main

import (
	"net/http"

	"bookstore/internal/handlers"
	"bookstore/internal/logic"
	"bookstore/internal/repository"
)

func RegisterRoutes(mux *http.ServeMux) {
	// Repository
	bookRepo := repository.NewBookRepo()

	// Service
	bookService := logic.NewBookService(bookRepo)

	// Handler
	bookHandler := handlers.NewBookHandler(bookService)

	// Routes
	mux.HandleFunc("/health", handlers.Health)
	mux.HandleFunc("/books", bookHandler.Books)     // GET, POST
	mux.HandleFunc("/books/", bookHandler.BookByID) // GET, PUT, DELETE
}
