package main

import (
 "net/http"

 "bookstore/internal/handlers"
 "bookstore/internal/logic"
 "bookstore/internal/repository"
)

func RegisterRoutes(mux *http.ServeMux) {
 bookRepo := repository.NewBookRepo()

 bookService := logic.NewBookService(bookRepo)

 bookHandler := handlers.NewBookHandler(bookService)

 mux.HandleFunc("/health", handlers.Health)
 mux.HandleFunc("/books", bookHandler.Books)     // GET, POST
 mux.HandleFunc("/books/", bookHandler.BookByID) // GET, PUT, DELETE
}
