package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/ASTeterin/loyalty/internal/model"
)

type orderRepo struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) model.OrderRepository {
	repo := &orderRepo{
		db: db,
	}
	return repo
}

func (repo *orderRepo) Store(order model.Order) error {
	fmt.Println("store", order)
	ctx := context.TODO()
	const query = `
        INSERT INTO orders (id, user_id, status, created_at)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (id) 
        DO UPDATE SET
    		status = EXCLUDED.status
    `

	_, err := repo.db.ExecContext(ctx, query, order.ID, order.UserID.String(), string(order.Status), order.CreatedAt)
	return err
}

func (repo *orderRepo) GetByOrderID(orderID int) (*model.Order, error) {
	ctx := context.TODO()
	query := `SELECT id, user_id, status, created_at FROM orders WHERE id = $1`
	order := model.Order{}
	var strOrder string
	err := repo.db.QueryRowContext(ctx, query, orderID).Scan(
		&order.ID, &order.UserID, &strOrder, &order.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, model.ErrOrderNotFound
	}
	order.Status = model.OrderStatus(strOrder)

	return &order, err
}

func (repo *orderRepo) ListOrders(userID string) ([]model.Order, error) {
	ctx := context.TODO()
	query := `SELECT id, user_id, status, created_at FROM orders WHERE user_id = $1 ORDER BY created_at DESC`
	urls := make([]model.Order, 0)
	rows, err := repo.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var order model.Order
		var strOrder string
		err = rows.Scan(&order.ID, &order.UserID, &strOrder, &order.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}
		order.Status = model.OrderStatus(strOrder)
		urls = append(urls, order)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return urls, nil
}
