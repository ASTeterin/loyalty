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

func (repo *orderRepo) Store(order model.Order) error {
	ctx := context.TODO()
	const query = `
        INSERT INTO orders (id, user_id, status, created_at)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (id) DO NOTHING
    `

	_, err := repo.db.ExecContext(ctx, query, order.ID, order.UserID.String(), order.Status, order.CreatedAt)
	return err
}

func (repo *orderRepo) GetByOrderID(orderID int) (*model.Order, error) {
	ctx := context.TODO()
	query := `SELECT id, user_id, status, created_at FROM orders WHERE id = $1`
	order := model.Order{}
	err := repo.db.QueryRowContext(ctx, query, orderID).Scan(
		&order.ID, &order.UserID, &order.Status, &order.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, model.ErrOrderNotFound
	}

	return &order, err
}
