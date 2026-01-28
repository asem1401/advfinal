package main

import (
	"log"
	"net/http"

	"bookstore/internal/logic"
)

func main() {
	logic.StartCartWorkerPool(3)

	mux := http.NewServeMux()
	RegisterRoutes(mux)

	log.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
