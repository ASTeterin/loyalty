package service

import (
	"errors"
	"golang.org/x/crypto/bcrypt"

	"github.com/ASTeterin/loyalty/internal/model"
)

const batchSize = 50

var (
	ErrUserExists = errors.New("user already exists")
)

type ShortenerService interface {
	Register(originalURL, userID string) (*string, error)
}

func NewShortenerService(repo model.UserRepository, maxWorkers int) ShortenerService {
	return &shortenerService{
		repo:       repo,
		maxWorkers: maxWorkers,
		batchSize:  batchSize,
	}
}

type shortenerService struct {
	repo       model.UserRepository
	batchSize  int
	maxWorkers int
}

func (s *shortenerService) Register(login, password string) (*string, error) {
	_, err := s.repo.GetByLogin(login)
	if err != nil {
		if errors.Is(err, model.ErrUserNotFound) {
			hash, err2 := HashPassword(password)
			if err2 != nil {
				return nil, err2
			}
			id, err2 := s.repo.NextUserID()
			if err2 != nil {
				return nil, err2
			}
			user := model.NewUser(id, login, hash)
			return s.repo.Store(user)
		}
	}
	return nil, ErrUserExists
}

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// Проверка пароля
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
