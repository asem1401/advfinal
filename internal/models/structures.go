package models

type User struct {
	ID       int
	Email    string
	Password string
}

type Book struct {
	ID    int
	Title string
	Price float64
}

type Order struct {
	ID     int
	UserID int
	Total  float64
}
