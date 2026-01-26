package main

import (
	"log"
	"net/http"

	"bookstore/internal/handlers"
	"bookstore/internal/logic"
)

func main() {
	logic.StartCartWorkerPool(3)

	mux := http.NewServeMux()

	RegisterRoutes(mux)

	mux.HandleFunc("/cart/add", handlers.AddToCartHandler)

	log.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
