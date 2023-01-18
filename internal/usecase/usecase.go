package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Spear5030/YAGopherMart/domain"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"io"
	"net/http"
	"time"
)

type Storager interface {
	RegisterUser(ctx context.Context, login string, hash string) (int, error)
	GetUserHash(ctx context.Context, login string) (int, string, error)
	PostOrder(ctx context.Context, num string, userID int) error
	GetOrders(ctx context.Context, userID int) ([]domain.Order, error)
	UpdateOrder(ctx context.Context, accrual domain.Accrual) error
	GetBalance(ctx context.Context, userID int) (float64, error)
	GetWithdrawn(ctx context.Context, userID int) (float64, error)
	PostWithdraw(ctx context.Context, userID int, order string, sum float64) error
	GetWithdrawals(ctx context.Context, userID int) ([]domain.Withdraw, error)
}

type usecase struct {
	logger     *zap.Logger
	storage    Storager
	accrualDSN string
}

func New(logger *zap.Logger, storage Storager, accDSN string) *usecase {
	return &usecase{
		logger:     logger,
		storage:    storage,
		accrualDSN: accDSN,
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
	return id, nil
}

func (uc *usecase) PostOrder(ctx context.Context, num string, userID int) error {
	err := uc.storage.PostOrder(ctx, num, userID)
	if err != nil {
		return err
	}
	return nil
}

func (uc *usecase) GetOrders(ctx context.Context, userID int) ([]domain.Order, error) {
	return uc.storage.GetOrders(ctx, userID)
}

func (uc *usecase) GetWithdrawals(ctx context.Context, userID int) ([]domain.Withdraw, error) {
	return uc.storage.GetWithdrawals(ctx, userID)
}

func (uc *usecase) WorkWithOrder(ctx context.Context, num string) error {
	//context.WithCancel(ctx)
	resp, err := http.Get(uc.accrualDSN + "/api/orders/" + num)
	// http.NewRequestWithContext(ctx, http.MethodGet, "https://example.com", nil)
	//http.DefaultClient.Do() //todo get with context
	if err != nil {
		uc.logger.Debug("workOrder error", zap.Error(err))
		return err
	}
	if resp.StatusCode == http.StatusOK {
		b, err := io.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err != nil {
			return err
		}

		var acc domain.Accrual
		if err = json.Unmarshal(b, &acc); err != nil {
			uc.logger.Debug("workOrder unmarshal error", zap.Error(err))
			return err
		}
		err = uc.storage.UpdateOrder(ctx, acc)
		if err != nil {
			uc.logger.Debug("workOrder update error", zap.Error(err))
			return err
		}
	} else {
		uc.logger.Debug("workWithOrder not 200", zap.String("code", resp.Status))
		if resp.StatusCode == http.StatusTooManyRequests {
			return errors.New("429 retry") //todo new error in domain
		}
		return errors.New("some error. won't work with order")
	}
	return nil
}

func (uc *usecase) GetBalance(ctx context.Context, userID int) (float64, error) {
	return uc.storage.GetBalance(ctx, userID)
}

func (uc *usecase) GetBalanceAndWithdrawn(ctx context.Context, userID int) (balance float64, withdrawn float64, err error) {
	balance, err = uc.storage.GetBalance(ctx, userID)
	uc.logger.Debug("uc get balance", zap.Float64("balance", balance))
	if err != nil {
		return 0, 0, err
	}
	withdrawn, err = uc.storage.GetWithdrawn(ctx, userID)
	if err != nil {
		return balance, 0, err
	}
	return
}

func (uc *usecase) PostWithdraw(ctx context.Context, userID int, order string, sum float64) error {
	balance, err := uc.storage.GetBalance(ctx, userID)
	if err != nil {
		return err
	}
	if balance < sum {
		return domain.ErrInsufficientBalance
	}
	uc.logger.Debug("uc post balance", zap.Float64("balance", balance))
	uc.logger.Debug("uc post sum", zap.Float64("sum", sum))
	return uc.storage.PostWithdraw(ctx, userID, order, sum)
}

func genToken(id int) (string, error) {
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256, jwt.MapClaims{
			"ID":        id,
			"ExpiresAt": jwt.NewNumericDate(time.Now().Add(time.Hour * 24)), // TODO Config?
		})
	tokenString, err := token.SignedString([]byte("hmacSampleSecret"))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
