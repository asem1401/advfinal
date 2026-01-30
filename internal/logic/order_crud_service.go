package logic

import (
  "errors"

  "bookstore/internal/models"
  "bookstore/internal/repository"
)

type OrderCRUDService struct {
  repo     repository.OrderRepository
  bookRepo repository.BookRepository
}

func NewOrderCRUDService(repo repository.OrderRepository, bookRepo repository.BookRepository) *OrderCRUDService {
  return &OrderCRUDService{repo: repo, bookRepo: bookRepo}
}

func (s *OrderCRUDService) CreateOrder(customerID int, items []models.OrderItem) (models.Order, []models.OrderItem, error) {
  if customerID <= 0 {
    customerID = 1
  }
  if len(items) == 0 {
    return models.Order{}, nil, errors.New("items required")
  }

  for _, it := range items {
    if _, err := s.bookRepo.GetByID(it.BookID); err != nil {
      return models.Order{}, nil, errors.New("book not found")
    }
  }

  return s.repo.Create(customerID, items)
}

func (s *OrderCRUDService) ListOrders() []models.Order {
  return s.repo.GetAll()
}

func (s *OrderCRUDService) GetOrder(id int) (models.Order, []models.OrderItem, error) {
  return s.repo.GetByID(id)
}

func (s *OrderCRUDService) UpdateOrder(o models.Order) error {
  if o.ID <= 0 {
    return errors.New("order id must be positive")
  }
  if o.Total < 0 {
    return errors.New("total cannot be negative")
  }
  return s.repo.Update(o)
}

func (s *OrderCRUDService) DeleteOrder(id int) error {
  return s.repo.Delete(id)
}
