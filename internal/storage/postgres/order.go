package postgres

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"go-pet-shop/internal/models"
	"go-pet-shop/internal/storage"
	"slices"

	"github.com/jackc/pgx/v5"
)

func (s *Storage) CreateOrder(ctx context.Context, order models.Order) (int, error) {
	const fn = "storage.postgres.order.CreateOrder"

	var id int
	err := s.db.QueryRow(ctx,
		`INSERT INTO orders (user_email, total_price)
		VALUES ($1, $2)
		RETURNING id`,
		order.UserEmail, order.TotalPrice).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	return id, nil
}

func (s *Storage) AddOrderItem(ctx context.Context, orderItem models.OrderItem) error {
	const fn = "storage.postgres.order.AddOrderItem"

	_, err := s.db.Exec(ctx,
		`INSERT INTO order_items (order_id, product_id, quantity)
		VALUES ($1, $2, $3)`,
		orderItem.OrderID, orderItem.ProductID, orderItem.Quantity)
	if err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}
	return nil
}

func (s *Storage) GetOrderByID(ctx context.Context, id int) (*models.Order, error) {
	const fn = "storage.postgres.order.GetOrderByID"

	var order models.Order
	err := s.db.QueryRow(ctx,
		`SELECT id, user_email, total_price, created_at
		FROM orders WHERE id = $1`,
		id).Scan(&order.ID, &order.UserEmail, &order.TotalPrice, &order.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w: id=%d", fn, storage.ErrNotFound, id)
		}
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	return &order, nil
}

func (s *Storage) GetOrdersByUserEmail(ctx context.Context, email string) ([]models.Order, error) {
	const fn = "storage.postgres.order.GetOrdersByUserEmail"

	var orders []models.Order
	rows, err := s.db.Query(ctx,
		`SELECT id, user_email, total_price, created_at
		FROM orders
		WHERE user_email = $1
		ORDER BY created_at DESC`,
		email)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	defer rows.Close()

	for rows.Next() {
		var order models.Order
		if err := rows.Scan(&order.ID, &order.UserEmail, &order.TotalPrice, &order.CreatedAt); err != nil {
			return nil, fmt.Errorf("%s: %w", fn, err)
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	return orders, nil
}

func (s *Storage) GetOrderItemsByOrderID(ctx context.Context, orderID int) ([]models.OrderItem, error) {
	const fn = "storage.postgres.order.GetOrderItemsByOrderID"

	var orderItems []models.OrderItem
	rows, err := s.db.Query(ctx,
		`SELECT id, order_id, product_id, quantity
		FROM order_items
		WHERE order_id = $1
		ORDER BY id`,
		orderID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	defer rows.Close()

	for rows.Next() {
		var orderItem models.OrderItem
		if err := rows.Scan(&orderItem.ID, &orderItem.OrderID, &orderItem.ProductID, &orderItem.Quantity); err != nil {
			return nil, fmt.Errorf("%s: %w", fn, err)
		}
		orderItems = append(orderItems, orderItem)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	return orderItems, nil
}

func (s *Storage) PlaceOrder(ctx context.Context, userEmail string, items []models.OrderItem) (orderID int, err error) {
	const fn = "storage.postgres.order.PlaceOrder"

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}
	defer tx.Rollback(ctx)

	sortedItems := slices.Clone(items)
	slices.SortFunc(sortedItems, func(a, b models.OrderItem) int {
		return cmp.Compare(a.ProductID, b.ProductID)
	})

	var totalPrice float64
	for _, item := range sortedItems {
		var price float64
		err = tx.QueryRow(ctx,
			`UPDATE products SET stock = stock - $1
			WHERE id = $2 AND stock >= $1
			RETURNING price`, item.Quantity, item.ProductID).Scan(&price)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				var exists bool
				err = tx.QueryRow(ctx,
					`SELECT EXISTS(SELECT 1 FROM products WHERE id = $1)`,
					item.ProductID).Scan(&exists)
				if err != nil {
					return 0, fmt.Errorf("%s: %w", fn, err)
				}
				if !exists {
					return 0, fmt.Errorf("%s: %w: product_id=%d", fn, storage.ErrNotFound, item.ProductID)
				}
				return 0, fmt.Errorf("%s: %w: product_id=%d", fn, storage.ErrInsufficientStock, item.ProductID)
			}
			return 0, fmt.Errorf("%s: %w", fn, err)
		}
		totalPrice += price * float64(item.Quantity)
	}

	err = tx.QueryRow(ctx,
		`INSERT INTO orders (user_email, total_price)
		VALUES ($1, $2)
		RETURNING id`, userEmail, totalPrice).Scan(&orderID)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	for _, item := range sortedItems {
		_, err = tx.Exec(ctx,
			`INSERT INTO order_items (order_id, product_id, quantity)
			VALUES ($1, $2, $3)`,
			orderID, item.ProductID, item.Quantity)
		if err != nil {
			return 0, fmt.Errorf("%s: %w", fn, err)
		}
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO transactions (order_id, amount, status)
		VALUES ($1, $2, 'pending')`, orderID, totalPrice)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}
	return orderID, nil
}

func (s *Storage) GetUserOrderHistory(ctx context.Context, userEmail string) ([]models.OrderDetail, error) {
	const fn = "storage.postgres.order.GetUserOrderHistory"

	rows, err := s.db.Query(ctx,
		`SELECT o.id, o.user_email, o.total_price, o.created_at, COALESCE(t.status, '') AS transaction_status,
		oi.id, oi.order_id, oi.product_id, oi.quantity
		FROM orders o
		JOIN order_items oi ON o.id = oi.order_id
		LEFT JOIN transactions t ON o.id = t.order_id
		WHERE o.user_email = $1
		ORDER BY o.created_at DESC, o.id DESC, oi.id`, userEmail)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	defer rows.Close()

	byID := make(map[int]*models.OrderDetail)
	result := make([]*models.OrderDetail, 0)

	for rows.Next() {
		var o models.Order
		var status string
		var item models.OrderItem
		if err := rows.Scan(&o.ID, &o.UserEmail, &o.TotalPrice, &o.CreatedAt, &status, &item.ID, &item.OrderID, &item.ProductID, &item.Quantity); err != nil {
			return nil, fmt.Errorf("%s: %w", fn, err)
		}
		od, ok := byID[o.ID]
		if !ok {
			od = &models.OrderDetail{
				Order:             o,
				Items:             []models.OrderItem{},
				TransactionStatus: status,
			}
			byID[o.ID] = od
			result = append(result, od)
		}
		od.Items = append(od.Items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	orderDetails := make([]models.OrderDetail, 0, len(result))
	for _, od := range result {
		orderDetails = append(orderDetails, *od)
	}
	return orderDetails, nil
}
