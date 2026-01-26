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
 return &BookHandler{
  service: service,
 }
}

func (h *BookHandler) Books(w http.ResponseWriter, r *http.Request) {
 switch r.Method {

 case http.MethodGet:
  books := h.service.ListBooks()
  writeJSON(w, http.StatusOK, books)

 case http.MethodPost:
  var in models.Book
  if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
   writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
   return
  }

  if err := h.service.CreateBook(in); err != nil {
   writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
   return
  }

  writeJSON(w, http.StatusCreated, map[string]string{"message": "created"})

 default:
  writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
 }
}

func (h *BookHandler) BookByID(w http.ResponseWriter, r *http.Request) {
 idStr := strings.TrimPrefix(r.URL.Path, "/books/")
 id, err := strconv.Atoi(idStr)
 if err != nil || id <= 0 {
  writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id must be positive int"})
  return
 }

 switch r.Method {

 case http.MethodGet:
  book, err := h.service.GetBook(id)
  if err != nil {
   writeJSON(w, http.StatusNotFound, map[string]string{"error": "book not found"})
   return
  }
  writeJSON(w, http.StatusOK, book)

 case http.MethodPut:
  var patch models.Book
  if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
   writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
   return
  }

  patch.ID = id
  if err := h.service.UpdateBook(patch); err != nil {
   writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
   return
  }

  writeJSON(w, http.StatusOK, map[string]string{"message": "updated"})

 case http.MethodDelete:
  if err := h.service.DeleteBook(id); err != nil {
   writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
   return
  }
  writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})

 default:
  writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
 }
}
