package repository

import (
	"errors"
	"sync"

	"bookstore/internal/models"
)

type BookRepository interface {
	Create(book models.Book) error
	GetByID(id int) (models.Book, error)
	GetAll() []models.Book
	Update(book models.Book) error
	Delete(id int) error
}

type BookRepo struct {
	mu    sync.RWMutex
	books map[int]models.Book
}

func NewBookRepo() *BookRepo {
	return &BookRepo{
		books: make(map[int]models.Book),
	}
}

func (r *BookRepo) Create(book models.Book) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.books[book.ID]; exists {
		return errors.New("book already exists")
	}

	r.books[book.ID] = book
	return nil
}

func (r *BookRepo) GetByID(id int) (models.Book, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	book, exists := r.books[id]
	if !exists {
		return models.Book{}, errors.New("book not found")
	}

	return book, nil
}

func (r *BookRepo) GetAll() []models.Book {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]models.Book, 0, len(r.books))
	for _, book := range r.books {
		result = append(result, book)
	}

	return result
}

func (r *BookRepo) Update(book models.Book) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.books[book.ID]; !exists {
		return errors.New("book not found")
	}

	r.books[book.ID] = book
	return nil
}

func (r *BookRepo) Delete(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.books[id]; !exists {
		return errors.New("book not found")
	}

	delete(r.books, id)
	return nil
}
