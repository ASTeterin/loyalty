package db

import (
	"context"
	"database/sql"
	"errors"
	"github.com/ASTeterin/loyalty/internal/model"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type userRepo struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) model.UserRepository {
	repo := &userRepo{
		db: db,
	}
	return repo
}

func (repo *userRepo) NextUserID() (uuid.UUID, error) {
	return uuid.NewV1()
}

func (repo *userRepo) Store(url model.User) (*string, error) {
	ctx := context.TODO()
	const query = `
        INSERT INTO user (id, login, password_hash)
        VALUES ($1, $2, $3)
        ON CONFLICT (id) DO NOTHING
        RETURNING short_url
    `

	var shortURL string
	err := repo.db.QueryRowContext(ctx, query, url.Short, url.Original, url.CreatedBy).Scan(&shortURL)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
		existingShortURL, err2 := repo.getStoredShortURL(url.Original)
		if err2 == nil {
			return &existingShortURL, model.ErrDuplicateURL
		}
		return nil, err
	}

	return &shortURL, nil
}

func (repo *userRepo) GetByLogin(login string) (*model.User, error) {
	ctx := context.TODO()
	query := `SELECT id, login, password_hash FROM user WHERE login = $1`
	user := model.User{}
	err := repo.db.QueryRowContext(ctx, query, login).Scan(
		&user.UUID, &user.Login, &user.PassHash)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, model.ErrUserNotFound
	}

	return &user, err
}
