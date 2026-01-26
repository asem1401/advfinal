package logic

import (
	"errors"

	"bookstore/internal/models"
	"bookstore/internal/repository"
)

func ListBooks() []models.Book {
	return repository.ListBooks()
}

func GetBook(id int) (models.Book, bool) {
	return repository.GetBook(id)
}

func CreateBook(b models.Book) (models.Book, error) {
	if b.BookID <= 0 {
		return models.Book{}, errors.New("book_id must be positive")
	}
	if b.Title == "" || b.Author == "" {
		return models.Book{}, errors.New("title and author required")
	}
	if b.Price < 0 {
		return models.Book{}, errors.New("price cannot be negative")
	}
	return repository.CreateBook(b)
}

func UpdateBook(id int, patch models.Book) (models.Book, error) {
	if patch.Price < 0 {
		return models.Book{}, errors.New("price cannot be negative")
	}
	return repository.UpdateBook(id, patch)
}
