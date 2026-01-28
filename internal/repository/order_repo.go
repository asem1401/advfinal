package repository

import (
  "errors"
  "sync"

  "bookstore/internal/models"
)

type OrderRepository interface {
  Create(customerID int, items []models.OrderItem) (models.Order, []models.OrderItem, error)
  GetByID(id int) (models.Order, []models.OrderItem, error)
  GetAll() []models.Order
  Update(order models.Order) error
  Delete(id int) error
}

type OrderRepo struct {
  mu sync.RWMutex

  nextOrderID int
  nextItemID  int

  orders     map[int]models.Order
  orderItems map[int][]models.OrderItem 
}

func NewOrderRepo() *OrderRepo {
  return &OrderRepo{
    nextOrderID: 1,
    nextItemID:  1,
    orders:      make(map[int]models.Order),
    orderItems:  make(map[int][]models.OrderItem),
  }
}

func (r *OrderRepo) Create(customerID int, items []models.OrderItem) (models.Order, []models.OrderItem, error) {
  r.mu.Lock()
  defer r.mu.Unlock()

  if len(items) == 0 {
    return models.Order{}, nil, errors.New("order items required")
  }

  var total float64
  outItems := make([]models.OrderItem, 0, len(items))
  for _, it := range items {
    if it.BookID <= 0 {
      return models.Order{}, nil, errors.New("bookId must be positive")
    }
    if it.Price < 0 {
      return models.Order{}, nil, errors.New("price cannot be negative")
    }
    it.ID = r.nextItemID
    r.nextItemID++
    total += it.Price
    outItems = append(outItems, it)
  }

  o := models.Order{
    ID:         r.nextOrderID,
    CustomerID: customerID,
    Total:      total,
  }
  r.nextOrderID++

  for i := range outItems {
    outItems[i].OrderID = o.ID
  }

  r.orders[o.ID] = o
  r.orderItems[o.ID] = outItems
  return o, append([]models.OrderItem(nil), outItems...), nil
}

func (r *OrderRepo) GetByID(id int) (models.Order, []models.OrderItem, error) {
  r.mu.RLock()
  defer r.mu.RUnlock()

  o, ok := r.orders[id]
  if !ok {
    return models.Order{}, nil, errors.New("order not found")
  }
  items := append([]models.OrderItem(nil), r.orderItems[id]...)
  return o, items, nil
}

func (r *OrderRepo) GetAll() []models.Order {
  r.mu.RLock()
  defer r.mu.RUnlock()

  out := make([]models.Order, 0, len(r.orders))
  for _, o := range r.orders {
    out = append(out, o)
  }
  return out
}

func (r *OrderRepo) Update(order models.Order) error {
  r.mu.Lock()
  defer r.mu.Unlock()

  if _, ok := r.orders[order.ID]; !ok {
    return errors.New("order not found")
  }
  r.orders[order.ID] = order
  return nil
}

func (r *OrderRepo) Delete(id int) error {
  r.mu.Lock()
  defer r.mu.Unlock()

  if _, ok := r.orders[id]; !ok {
    return errors.New("order not found")
  }
  delete(r.orders, id)
  delete(r.orderItems, id)
  return nil
}
