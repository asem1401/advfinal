package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	RegisterRoutes(mux)
	log.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
