package main

import (
	"net/http"

	"bookstore/internal/handlers"
)

func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", handlers.Health)
	mux.HandleFunc("/books", handlers.Books)     // GET, POST
	mux.HandleFunc("/books/", handlers.BookByID) // GET, PUT
}
