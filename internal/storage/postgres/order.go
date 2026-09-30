package postgres

import (
	"context"
	"fmt"
	"go-pet-shop/internal/models"
)

func (s *Storage) CreateOrder(ctx context.Context, order models.Order) (int, error) {
	const fn = "storage.postgres.order.CreateOrder"

	var id int
	err := s.db.QueryRow(ctx,
		`INSERT INTO orders (user_email, total_price) VALUES ($1, $2) RETURNING id`,
		order.UserEmail, order.TotalPrice).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	return id, nil
}

func (s *Storage) AddOrderItem(ctx context.Context, orderItem models.OrderItem) error {
	const fn = "storage.postgres.order.AddOrderItem"
	return fmt.Errorf("%s: not implemented", fn)
}

func (s *Storage) GetOrderByID(ctx context.Context, id int) (*models.Order, error) {
	const fn = "storage.postgres.order.GetOrderByID"
	return nil, fmt.Errorf("%s: not implemented", fn)
}

func (s *Storage) GetOrdersByUserEmail(ctx context.Context, email string) ([]models.Order, error) {
	const fn = "storage.postgres.order.GetOrdersByUserEmail"
	return nil, fmt.Errorf("%s: not implemented", fn)
}

func (s *Storage) GetOrderItemsByOrderID(ctx context.Context, orderID int) ([]models.OrderItem, error) {
	const fn = "storage.postgres.order.GetOrderItemsByOrderID"
	return nil, fmt.Errorf("%s: not implemented", fn)
}
