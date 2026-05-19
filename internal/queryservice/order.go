package query

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type OrderQueryService interface {
	ListOrders(userID string) ([]OrderDTO, error)
}

type OrderDTO struct {
	Number     string
	Status     string
	Accrual    float64
	UploadedAt time.Time
}

type orderQueryService struct {
	db *sql.DB
}

func NewOrderQueryService(db *sql.DB) OrderQueryService {
	q := &orderQueryService{
		db: db,
	}
	return q
}

func (repo *orderQueryService) ListOrders(userID string) ([]OrderDTO, error) {
	ctx := context.TODO()
	query := `SELECT o.id, o.status, t.points, o.created_at FROM orders o
				INNER JOIN balance_transaction t ON o.id = t.order_id
   				WHERE o.user_id = $1`

	orders := []OrderDTO{}
	rows, err := repo.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var o OrderDTO
		err = rows.Scan(&o.Number, &o.Status, &o.Accrual, &o.UploadedAt)
		if err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}
		orders = append(orders, o)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return orders, nil
}
