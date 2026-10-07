package models

import "time"

type Product struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
	Stock int     `json:"stock"` // количество на складе
}

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Order struct {
	ID         int       `json:"id"`
	UserEmail  string    `json:"user_email"`
	TotalPrice float64   `json:"total_price"`
	CreatedAt  time.Time `json:"created_at"`
}

type OrderItem struct {
	ID        int `json:"id"`
	OrderID   int `json:"order_id"`
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

type Transactions struct {
	ID        int       `json:"id"`
	OrderID   int       `json:"order_id"`
	Amount    float64   `json:"amount"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type OrderDetail struct {
	Order             Order       `json:"order"`
	TransactionStatus string      `json:"transaction_status"`
	Items             []OrderItem `json:"order_items"`
}

type PopularProduct struct {
	ProductID int    `json:"product_id"`
	Name      string `json:"name"`
	TotalSold int    `json:"total_sold"`
}
