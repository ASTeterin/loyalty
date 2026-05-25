package db

import (
	"context"
	"database/sql"
	"errors"

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

func (repo *orderRepo) Store(ctx context.Context, order model.Order) error {
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

func (repo *orderRepo) GetByOrderID(ctx context.Context, orderID string) (*model.Order, error) {
	query := `SELECT id, user_id, status, created_at FROM orders WHERE id = $1`
	order := model.Order{}
	var status string
	err := repo.db.QueryRowContext(ctx, query, orderID).Scan(
		&order.ID, &order.UserID, &status, &order.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, model.ErrOrderNotFound
	}
	order.Status = model.OrderStatus(status)

	return &order, err
}
