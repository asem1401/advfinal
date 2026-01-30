package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"bookstore/internal/logic"
	"bookstore/internal/models"
)

type CartHandler struct {
	service *logic.CartCRUDService
}

func NewCartHandler(service *logic.CartCRUDService) *CartHandler {
	return &CartHandler{service: service}
}

func (h *CartHandler) Carts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, h.service.ListCarts())

	case http.MethodPost:
		var in struct {
			CustomerID int `json:"customerId"`
		}
		_ = json.NewDecoder(r.Body).Decode(&in)
		c := h.service.CreateCart(in.CustomerID)
		writeJSON(w, http.StatusCreated, c)

	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

func (h *CartHandler) CartByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/carts/")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	switch r.Method {
	case http.MethodGet:
		c, items, err := h.service.GetCart(id)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"cart": c, "items": items})

	case http.MethodPut:
		var in models.Cart
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
			return
		}
		in.ID = id
		if err := h.service.UpdateCart(in); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "updated"})

	case http.MethodDelete:
		if err := h.service.DeleteCart(id); err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})

	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

func (h *CartHandler) CartItems(w http.ResponseWriter, r *http.Request) {
	// ожидаем /carts/{id}/items
	path := strings.TrimPrefix(r.URL.Path, "/carts/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "items" {
		http.NotFound(w, r)
		return
	}

	cartID, err := strconv.Atoi(parts[0])
	if err != nil || cartID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid cart id"})
		return
	}

	switch r.Method {
	case http.MethodPost:
		var in struct {
			BookID int `json:"bookId"`
			Qty    int `json:"qty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
			return
		}
		item, err := h.service.AddItem(cartID, in.BookID, in.Qty)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusCreated, item)

	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

func (h *CartHandler) CartItemByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/carts/")
	parts := strings.Split(path, "/")
	if len(parts) < 3 || parts[1] != "items" {
		http.NotFound(w, r)
		return
	}

	cartID, err := strconv.Atoi(parts[0])
	if err != nil || cartID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid cart id"})
		return
	}

	itemID, err := strconv.Atoi(parts[2])
	if err != nil || itemID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid item id"})
		return
	}

	switch r.Method {
	case http.MethodPut:
		var in struct {
			Qty int `json:"qty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
			return
		}
		if err := h.service.UpdateItem(cartID, itemID, in.Qty); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "updated"})

	case http.MethodDelete:
		if err := h.service.DeleteItem(cartID, itemID); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})

	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}
