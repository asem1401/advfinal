package repository

import (
	"errors"
	"sync"

	"bookstore/internal/models"
)

type WishlistRepository interface {
	Create(customerID int) models.Wishlist
	GetAll() []models.Wishlist
	GetByID(id int) (models.Wishlist, []models.WishlistItem, error)
	Delete(id int) error

	AddItem(wishlistID int, bookID int, qty int) (models.WishlistItem, error)
	DeleteItem(wishlistID int, itemID int) error
}

type WishlistRepo struct {
	mu sync.RWMutex

	nextWishlistID int
	nextItemID     int
	wishlists      map[int]models.Wishlist
	items          map[int][]models.WishlistItem
}

func NewWishlistRepo() *WishlistRepo {
	return &WishlistRepo{
		nextWishlistID: 1,
		nextItemID:     1,
		wishlists:      make(map[int]models.Wishlist),
		items:          make(map[int][]models.WishlistItem),
	}
}

func (r *WishlistRepo) Create(customerID int) models.Wishlist {
	r.mu.Lock()
	defer r.mu.Unlock()

	w := models.Wishlist{
		ID:         r.nextWishlistID,
		CustomerID: customerID,
	}
	r.nextWishlistID++
	r.wishlists[w.ID] = w
	r.items[w.ID] = []models.WishlistItem{}
	return w
}

func (r *WishlistRepo) GetAll() []models.Wishlist {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]models.Wishlist, 0, len(r.wishlists))
	for _, w := range r.wishlists {
		out = append(out, w)
	}
	return out
}

func (r *WishlistRepo) GetByID(id int) (models.Wishlist, []models.WishlistItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	w, ok := r.wishlists[id]
	if !ok {
		return models.Wishlist{}, nil, errors.New("wishlist not found")
	}
	items := append([]models.WishlistItem(nil), r.items[id]...)
	return w, items, nil
}

func (r *WishlistRepo) Delete(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.wishlists[id]; !ok {
		return errors.New("wishlist not found")
	}
	delete(r.wishlists, id)
	delete(r.items, id)
	return nil
}

func (r *WishlistRepo) AddItem(wishlistID int, bookID int, qty int) (models.WishlistItem, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.wishlists[wishlistID]; !ok {
		return models.WishlistItem{}, errors.New("wishlist not found")
	}
	if qty <= 0 {
		return models.WishlistItem{}, errors.New("qty must be > 0")
	}

	items := r.items[wishlistID]
	for i := range items {
		if items[i].BookID == bookID {
			items[i].Qty += qty
			r.items[wishlistID] = items
			return items[i], nil
		}
	}

	it := models.WishlistItem{
		ID:         r.nextItemID,
		WishlistID: wishlistID,
		BookID:     bookID,
		Qty:        qty,
	}
	r.nextItemID++
	r.items[wishlistID] = append(r.items[wishlistID], it)
	return it, nil
}

func (r *WishlistRepo) DeleteItem(wishlistID int, itemID int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	items := r.items[wishlistID]
	out := make([]models.WishlistItem, 0, len(items))
	found := false

	for _, it := range items {
		if it.ID == itemID {
			found = true
			continue
		}
		out = append(out, it)
	}
	if !found {
		return errors.New("item not found")
	}
	r.items[wishlistID] = out
	return nil
}
