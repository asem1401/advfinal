package main

import (
	"net/http"
	"strings"

	"bookstore/internal/handlers"
	"bookstore/internal/logic"
	"bookstore/internal/repository"
)

func RegisterRoutes(mux *http.ServeMux) {
	bookRepo := repository.NewBookRepo()
	cartRepo := repository.NewCartRepo()
	orderRepo := repository.NewOrderRepo()

	bookService := logic.NewBookService(bookRepo)
	cartCRUDService := logic.NewCartCRUDService(cartRepo, bookRepo)
	orderCRUDService := logic.NewOrderCRUDService(orderRepo, bookRepo)

	bookHandler := handlers.NewBookHandler(bookService)
	cartHandler := handlers.NewCartHandler(cartCRUDService)
	orderHandler := handlers.NewOrderHandler(orderCRUDService)

	mux.HandleFunc("/health", handlers.Health)

	mux.HandleFunc("/books", bookHandler.Books)    
	mux.HandleFunc("/books/", bookHandler.BookByID) 

	mux.HandleFunc("/carts", cartHandler.Carts) 

	mux.HandleFunc("/carts/", func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/items/") {
			cartHandler.CartItemByID(w, r)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/items") {
			cartHandler.CartItems(w, r)
			return
		}
		cartHandler.CartByID(w, r)
	})

	mux.HandleFunc("/orders", orderHandler.Orders)    
	mux.HandleFunc("/orders/", orderHandler.OrderByID) 
	mux.HandleFunc("/cart/add", handlers.AddToCartHandler)
	mux.HandleFunc("/login", handlers.Login)
}
