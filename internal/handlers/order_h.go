package handlers

import (
  "encoding/json"
  "net/http"
  "strconv"
  "strings"

  "bookstore/internal/logic"
  "bookstore/internal/models"
)

type OrderHandler struct {
  service *logic.OrderCRUDService
}

func NewOrderHandler(service *logic.OrderCRUDService) *OrderHandler {
  return &OrderHandler{service: service}
}

func (h *OrderHandler) Orders(w http.ResponseWriter, r *http.Request) {
  switch r.Method {
  case http.MethodGet:
    writeJSON(w, http.StatusOK, h.service.ListOrders())

  case http.MethodPost:
    var in struct {
      CustomerID int                `json:"customerId"`
      Items      []models.OrderItem `json:"items"`
    }
    if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
      writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
      return
    }
    o, items, err := h.service.CreateOrder(in.CustomerID, in.Items)
    if err != nil {
      writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
      return
    }
    writeJSON(w, http.StatusCreated, map[string]any{"order": o, "items": items})

  default:
    writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
  }
}

func (h *OrderHandler) OrderByID(w http.ResponseWriter, r *http.Request) {
  idStr := strings.TrimPrefix(r.URL.Path, "/orders/")
  id, err := strconv.Atoi(idStr)
  if err != nil || id <= 0 {
    writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
    return
  }

  switch r.Method {
  case http.MethodGet:
    o, items, err := h.service.GetOrder(id)
    if err != nil {
      writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
      return
    }
    writeJSON(w, http.StatusOK, map[string]any{"order": o, "items": items})

  case http.MethodPut:
    var in models.Order
    if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
      writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
      return
    }
    in.ID = id
    if err := h.service.UpdateOrder(in); err != nil {
      writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
      return
    }
    writeJSON(w, http.StatusOK, map[string]string{"message": "updated"})

  case http.MethodDelete:
    if err := h.service.DeleteOrder(id); err != nil {
      writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
      return
    }
    writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})

  default:
    writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
  }
}
