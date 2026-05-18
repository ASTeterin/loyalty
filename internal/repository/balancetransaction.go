package db

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/ASTeterin/loyalty/internal/model"
)

type balanceTransactionRepo struct {
	db *sql.DB
}

func NewBalanceTransactionRepository(db *sql.DB) model.BalanceTransactionRepository {
	repo := &balanceTransactionRepo{
		db: db,
	}
	return repo
}

func (repo *balanceTransactionRepo) Store(t model.BalanceTransaction) error {
	fmt.Println("store", t)
	ctx := context.TODO()
	const query = `
        INSERT INTO balance_transaction (order_id, user_id, points)
        VALUES ($1, $2, $3)
		ON CONFLICT (id) DO NOTHING
    `

	_, err := repo.db.ExecContext(ctx, query, t.OrderID, t.UserID.String(), t.Points)
	return err
}

func (repo *balanceTransactionRepo) ListUserTransactions(userID string) ([]model.BalanceTransaction, error) {
	ctx := context.TODO()
	query := `SELECT id, order_id, user_id, points, created_at FROM balance_transaction WHERE user_id = $1`
	transactions := make([]model.BalanceTransaction, 0)
	rows, err := repo.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var t model.BalanceTransaction
		err = rows.Scan(&t.ID, t.OrderID, &t.UserID, &t.Points, &t.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}
		transactions = append(transactions, t)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return transactions, nil
}
