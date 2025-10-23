package user

import (
	"context"
	"errors"
	"fmt"
	"notification/models"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrTokenGeneration    = errors.New("failed to generate token")
)

type Service struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Service { return &Service{db: db} }

type SignupRequest struct {
	Name     string
	Email    string
	Password string
}

type LoginRequest struct {
	Email    string
	Password string
}

func (s *Service) Signup(ctx context.Context, req SignupRequest) error {
	u := models.User{Name: req.Name, Email: req.Email}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	u.Password = string(hashedPassword)

	return s.db.WithContext(ctx).Create(&u).Error
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (string, error) {
	var u models.User
	if err := s.db.WithContext(ctx).Where("email = ?", req.Email).First(&u).Error; err != nil {
		return "", ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password)); err != nil {
		return "", ErrInvalidCredentials
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": u.ID,
		"exp": time.Now().Add(time.Hour * 24 * 30).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrTokenGeneration, err)
	}

	return tokenString, nil
}

func (s *Service) GetById(ctx context.Context, id uint) (models.User, error) {
	var u models.User
	if err := s.db.WithContext(ctx).First(&u, id).Error; err != nil {
		return models.User{}, fmt.Errorf("failed to find user: %w", err)
	}
	return u, nil
}
