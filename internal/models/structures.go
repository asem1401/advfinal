package models

<<<<<<< HEAD
import "time"

type Book struct {
	ID          int
	Title       string
	Author      string
	Genre       string
	Price       float64
	Description string
}

type Customer struct {
	ID       int
	Email    string
	Password string
	Address  string
}

type Cart struct {
	ID         int
	CustomerID int
	CreatedAt  time.Time
}

type CartItem struct {
	ID     int
	CartID int
	BookID int
	Qty    int
}

type Order struct {
	ID         int
	CustomerID int
	Total      float64
}

type OrderItem struct {
	ID      int
	OrderID int
	BookID  int
	Price   float64
}

type Payment struct {
	ID      int
	OrderID int
	Total   float64
	Status  string
=======
type User struct {
	ID       int
	Email    string
	Password string
}
type Book struct {
	BookID      int     `json:"book_id"`
	Title       string  `json:"title"`
	Author      string  `json:"author"`
	Genre       string  `json:"genre"`
	Price       float64 `json:"price"`
	Description string  `json:"description"`
}

type Order struct {
	ID     int
	UserID int
	Total  float64
>>>>>>> origin/main
}
