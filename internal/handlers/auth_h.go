package handlers

import (
	"net/http"
	"project/internal/logic"
)

func Login(w http.ResponseWriter, r *http.Request) {
	if logic.Login("test@mail.com", "1234") {
		w.Write([]byte("Login successful"))
	} else {
		w.WriteHeader(http.StatusUnauthorized)
	}
}
