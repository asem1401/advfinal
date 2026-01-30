package logic

import (
	"errors"

	"bookstore/internal/models"
	"bookstore/internal/repository"
)

// BookService handles business logic for books
type BookService struct {
	repo repository.BookRepository
}

// NewBookService creates a new BookService
func NewBookService(repo repository.BookRepository) *BookService {
	return &BookService{
		repo: repo,
	}
}

// ListBooks returns all books
func (s *BookService) ListBooks() []models.Book {
	return s.repo.GetAll()
}

// GetBook returns a book by ID
func (s *BookService) GetBook(id int) (models.Book, error) {
	return s.repo.GetByID(id)
}

// CreateBook validates and creates a new book
func (s *BookService) CreateBook(b models.Book) error {
	if b.ID <= 0 {
		return errors.New("book id must be positive")
	}
	if b.Title == "" || b.Author == "" {
		return errors.New("title and author are required")
	}
	if b.Price < 0 {
		return errors.New("price cannot be negative")
	}

	return s.repo.Create(b)
}

// UpdateBook validates and updates an existing book
func (s *BookService) UpdateBook(b models.Book) error {
	if b.Price < 0 {
		return errors.New("price cannot be negative")
	}

	return s.repo.Update(b)
}

// DeleteBook removes a book by ID
func (s *BookService) DeleteBook(id int) error {
	return s.repo.Delete(id)
}
