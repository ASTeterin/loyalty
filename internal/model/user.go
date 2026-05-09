package model

import (
	"errors"
	"github.com/gofrs/uuid"
)

var (
	ErrUserNotFound = errors.New("url not found")
	ErrDuplicateURL = errors.New("duplicate url")
)

const (
	ShortURLLen = 8
	Letters     = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

type BatchDeleteResult struct {
	SuccessCount int
	Error        error
}

type User struct {
	UUID     uuid.UUID `json:"uuid" db:"id"`
	Login    string    `json:"login" db:"login"`
	PassHash string    `json:"pass" db:"pass_hash"`
}

type UserRepository interface {
	NextUserID() (uuid.UUID, error)
	Store(user User) (*string, error)
	GetByLogin(login string) (*User, error)
}

func NewUser(id uuid.UUID, login, passHash string) User {
	return User{
		UUID:     id,
		Login:    login,
		PassHash: passHash,
	}
}
