package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"bookstore/internal/logic"
	"bookstore/internal/models"
)

type BookHandler struct {
	service *logic.BookService
}

func NewBookHandler(service *logic.BookService) *BookHandler {
	return &BookHandler{service: service}
}

func (h *BookHandler) Books(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		q := parseBookQuery(r)
		books, err := h.service.ListBooks(r.Context(), q)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "internal error"})
			return
		}
		_ = json.NewEncoder(w).Encode(books)

	case http.MethodPost:
		var b models.Book
		if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid json"})
			return
		}

		created, err := h.service.CreateBook(b)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(created)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
	}
}

func (h *BookHandler) BookByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid id"})
		return
	}

	switch r.Method {
	case http.MethodGet:
		b, err := h.service.GetBook(id)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		_ = json.NewEncoder(w).Encode(b)

	case http.MethodPut:
		var b models.Book
		if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid json"})
			return
		}

		b.ID = id
		if err := h.service.UpdateBook(b); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		_ = json.NewEncoder(w).Encode(b)

	case http.MethodDelete:
		if err := h.service.DeleteBook(id); err != nil {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
	}
}

func parseBookQuery(r *http.Request) models.BookQuery {
	qp := r.URL.Query()

	var q models.BookQuery
	q.Genre = strings.TrimSpace(qp.Get("genre"))
	q.SortBy = strings.TrimSpace(qp.Get("sort"))
	q.Order = strings.TrimSpace(qp.Get("order"))

	if v := qp.Get("minPrice"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			q.MinPrice = &f
		}
	}
	if v := qp.Get("maxPrice"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			q.MaxPrice = &f
		}
	}

	return q
}
