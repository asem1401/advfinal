package main

import (
	"bookstore/internal/middleware"
	"log"
	"net/http"
	"os"
	"strings"

	"bookstore/internal/handlers"
	"bookstore/internal/logic"
	"bookstore/internal/repository"

	"go.mongodb.org/mongo-driver/mongo"
)

func RegisterRoutes(mux *http.ServeMux, mongoDB *mongo.Database) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT_SECRET is not set")
	}

	bookRepo := repository.NewBookRepo(mongoDB)
	userRepo := repository.NewUserRepo(mongoDB)
	cartRepo := repository.NewCartRepo()
	wishlistRepo := repository.NewWishlistRepo(mongoDB)

	logic.StartOrderWorkerPool(2, cartRepo, wishlistRepo)

	orderRepo := repository.NewOrderRepo(mongoDB)

	bookService := logic.NewBookService(bookRepo)
	cartCRUDService := logic.NewCartCRUDService(cartRepo, bookRepo)

	orderSvc := logic.NewOrderService(orderRepo, bookRepo, cartRepo)
	orderCRUD := logic.NewOrderCRUDService(orderRepo)

	wishlistService := logic.NewWishlistService(wishlistRepo, bookRepo, orderRepo)
	authService := logic.NewAuthService(userRepo, secret)

	bookHandler := handlers.NewBookHandler(bookService)
	cartHandler := handlers.NewCartHandler(cartCRUDService)

	orderHandler := handlers.NewOrderHandler(orderSvc)
	orderCRUDHandler := handlers.NewOrderCRUDHandler(orderCRUD)

	wishlistHandler := handlers.NewWishlistHandler(wishlistService)
	authHandler := handlers.NewAuthHandler(authService)

	mux.HandleFunc("/health", handlers.Health)

	mux.HandleFunc("/auth/register", authHandler.Register)
	mux.HandleFunc("/auth/login", authHandler.Login)

	mux.HandleFunc("/books", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			bookHandler.Books(w, r)
		case http.MethodPost:
			middleware.AdminOnly(secret, bookHandler.Books)(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/books/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			bookHandler.BookByID(w, r)
		case http.MethodPut, http.MethodDelete:
			middleware.AdminOnly(secret, bookHandler.BookByID)(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/carts", middleware.AuthOnly(secret, cartHandler.Carts))
	mux.HandleFunc("/carts/", middleware.AuthOnly(secret, func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/items/") {
			cartHandler.CartItemByID(w, r)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/items") {
			cartHandler.CartItems(w, r)
			return
		}
		cartHandler.CartByID(w, r)
	}))

	mux.HandleFunc("/orders", middleware.AuthOnly(secret, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			orderHandler.Orders(w, r)
			return
		}
		orderCRUDHandler.Orders(w, r)
	}))

	mux.HandleFunc("/orders/", middleware.AuthOnly(secret, orderCRUDHandler.OrderByID))

	mux.HandleFunc("/wishlists", wishlistHandler.Wishlists)
	mux.HandleFunc("/wishlists/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/items") {
			wishlistHandler.WishlistItems(w, r)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/gift") {
			wishlistHandler.Gift(w, r)
			return
		}
		wishlistHandler.WishlistByID(w, r)
	})

	mux.HandleFunc("/admin/ping", middleware.AdminOnly(secret, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"admin ok"}`))
	}))

	mux.HandleFunc("/cart/add", handlers.AddToCartHandler)
}
