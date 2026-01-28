package logic

import (
	"errors"

	"bookstore/internal/models"
	"bookstore/internal/repository"
)

type WishlistService struct {
	wRepo     repository.WishlistRepository
	bookRepo  repository.BookRepository
	orderRepo repository.OrderRepository
}

func NewWishlistService(
	wRepo repository.WishlistRepository,
	bookRepo repository.BookRepository,
	orderRepo repository.OrderRepository,
) *WishlistService {
	return &WishlistService{
		wRepo:     wRepo,
		bookRepo:  bookRepo,
		orderRepo: orderRepo,
	}
}

func (s *WishlistService) CreateWishlist(customerID int) models.Wishlist {
	if customerID <= 0 {
		customerID = 1
	}
	return s.wRepo.Create(customerID)
}

func (s *WishlistService) ListWishlists() []models.Wishlist {
	return s.wRepo.GetAll()
}

func (s *WishlistService) GetWishlist(id int) (models.Wishlist, []models.WishlistItem, error) {
	return s.wRepo.GetByID(id)
}

func (s *WishlistService) AddItem(wishlistID, bookID, qty int) (models.WishlistItem, error) {
	if _, err := s.bookRepo.GetByID(bookID); err != nil {
		return models.WishlistItem{}, errors.New("book not found")
	}
	return s.wRepo.AddItem(wishlistID, bookID, qty)
}

func (s *WishlistService) GiftFromWishlist(wishlistID int, buyerID int) (models.Order, []models.OrderItem, int, error) {
	if buyerID <= 0 {
		return models.Order{}, nil, 0, errors.New("buyerCustomerId must be positive")
	}

	w, items, err := s.wRepo.GetByID(wishlistID)
	if err != nil {
		return models.Order{}, nil, 0, err
	}
	if len(items) == 0 {
		return models.Order{}, nil, 0, errors.New("wishlist is empty")
	}

	var orderItems []models.OrderItem
	for _, wi := range items {
		book, err := s.bookRepo.GetByID(wi.BookID)
		if err != nil {
			return models.Order{}, nil, 0, errors.New("book not found")
		}
		
		orderItems = append(orderItems, models.OrderItem{
			BookID: wi.BookID,
			Price:  book.Price,
		})
	}

	order, createdItems, err := s.orderRepo.Create(buyerID, orderItems)
	if err != nil {
		return models.Order{}, nil, 0, err
	}

	return order, createdItems, w.CustomerID, nil
}
