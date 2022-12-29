package usecase

import (
	"context"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type storager interface {
	RegisterUser(ctx context.Context, login string, hash string) (int, error)
}

type usecase struct {
	logger  *zap.Logger
	storage storager
}

func New(logger *zap.Logger, storage storager) *usecase {
	return &usecase{
		logger:  logger,
		storage: storage,
	}
}

func (uc *usecase) RegisterUser(ctx context.Context, login string, password string) (int, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}
	id, err := uc.storage.RegisterUser(ctx, login, string(hashedPassword))
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256, jwt.MapClaims{
			"id":        id,
			"ExpiresAt": jwt.NewNumericDate(time.Unix(1516239022, 0)),
		})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString("hmacSampleSecret")

	return id, err
}
