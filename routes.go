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
 wishlistRepo := repository.NewWishlistRepo()
 userRepo := repository.NewUserRepo()

 
 bookService := logic.NewBookService(bookRepo)
 cartCRUDService := logic.NewCartCRUDService(cartRepo, bookRepo)
 orderCRUDService := logic.NewOrderCRUDService(orderRepo, bookRepo)
 wishlistService := logic.NewWishlistService(wishlistRepo, bookRepo, orderRepo)
 authService := logic.NewAuthService(userRepo, "SECRET_KEY")

 
 bookHandler := handlers.NewBookHandler(bookService)
 cartHandler := handlers.NewCartHandler(cartCRUDService)
 orderHandler := handlers.NewOrderHandler(orderCRUDService)
 wishlistHandler := handlers.NewWishlistHandler(wishlistService)
 authHandler := handlers.NewAuthHandler(authService)

 
 mux.HandleFunc("/health", handlers.Health)

 
 mux.HandleFunc("/auth/register", authHandler.Register)
 mux.HandleFunc("/auth/login", authHandler.Login)

 
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
}
