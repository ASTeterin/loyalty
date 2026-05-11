package service

import (
	"errors"
	"golang.org/x/crypto/bcrypt"

	"github.com/ASTeterin/loyalty/internal/model"
)

var (
	ErrUserExists = errors.New("user already exists")
)

type UserService interface {
	Register(originalURL, userID string) (*string, error)
}

func NewUserService(repo model.UserRepository) UserService {
	return &userService{
		repo: repo,
	}
}

type userService struct {
	repo model.UserRepository
}

func (s *userService) Register(login, password string) (*string, error) {
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
			err2 = s.repo.Store(user)
			strID := id.String()
			return &strID, err2
		}
		return nil, err
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
