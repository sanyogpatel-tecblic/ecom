package models

import "database/sql"

type Category struct {
	ID         int    `json:"id" validate:"required"`
	Category   string `json:"category" validate:"required"`
	Statuscode int    `json:"status"`
	ImageURL   string `json:"imageurl"`
}

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Product struct {
	ID             int      `json:"id" validate:"required"`
	Name           string   `json:"name" validate:"required"`
	CategoryID     int      `json:"category_id"`
	Description    string   `json:"description" validate:"required"`
	ImageURL       string   `json:"imageurl"`
	Seller         string   `json:"seller" validate:"required"`
	Price          int      `json:"price" validate:"required"`
	Highlights     string   `json:"highlights" validate:"required"`
	Specifications string   `json:"specifications" validate:"required"`
	Category       Category `json:"category" validate:"required"`
}
type User struct {
	ID       int            `json:"id" validate:"required"`
	Username string         `json:"username" validate:"required"`
	Password string         `json:"password" validate:"required"`
	Email    string         `json:"email" validate:"required"`
	Name     sql.NullString `json:"name" validate:"required"`
	Gender   sql.NullString `json:"gender" validate:"required"`
	Mobile   sql.NullString `json:"mobile" validate:"required"`
	// Orders   string `json:"orders" validate:"required"`
	ImageURL sql.NullString `json:"imageurl" validate:"required"`
	Address  sql.NullString `json:"address" validate:"required"`
}
type Cart struct {
	ID          int `json:"id" validate:"required"`
	UserID      int `json:"user_id"`
	ProductID   int `json:"product_id"`
	Quantity    int `json:"quantity"`
	Final_price int `json:"final_price"`
	Product     Product
}
