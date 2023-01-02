package usecase

import (
	"context"
	"github.com/Spear5030/YAGopherMart/domain"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type storager interface {
	RegisterUser(ctx context.Context, login string, hash string) (int, error)
	GetUserHash(ctx context.Context, login string) (int, string, error)
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
	if err != nil {
		return 0, err
	}

	//return genToken(id)
	return id, nil
}

func (uc *usecase) LoginUser(ctx context.Context, login string, password string) (int, error) {

	id, hash, err := uc.storage.GetUserHash(ctx, login)
	if err != nil {
		return 0, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return 0, domain.ErrInvalidPassword
	}
	//return genToken(id)
	return id, nil
}

func genToken(id int) (string, error) {
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256, jwt.MapClaims{
			"ID":        id,
			"ExpiresAt": jwt.NewNumericDate(time.Now().Add(time.Hour * 24)), // TODO Config?
		})
	tokenString, err := token.SignedString([]byte("hmacSampleSecret")) //TODO Config
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
