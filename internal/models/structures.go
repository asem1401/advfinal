package models

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
}
