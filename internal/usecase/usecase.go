package usecase

import (
	"context"
	"encoding/json"
	"github.com/Spear5030/YAGopherMart/domain"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"io"
	"net/http"
	"time"
)

type storager interface {
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
	storage    storager
	accrualDSN string
}

func New(logger *zap.Logger, storage storager, accDSN string) *usecase {
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

func (uc *usecase) PostOrder(ctx context.Context, num string, userID int) error {
	err := uc.storage.PostOrder(ctx, num, userID)
	if err != nil {
		return err
	}
	//time.AfterFunc(500*time.Millisecond, func() { uc.WorkWithOrder(ctx, num) })
	return nil
}

func (uc *usecase) GetOrders(ctx context.Context, userID int) ([]domain.Order, error) {
	return uc.storage.GetOrders(ctx, userID)
}

func (uc *usecase) GetWithdrawals(ctx context.Context, userID int) ([]domain.Withdraw, error) {
	return uc.storage.GetWithdrawals(ctx, userID)
}

func (uc *usecase) WorkWithOrder(ctx context.Context, num string) {
	resp, err := http.Get(uc.accrualDSN + "/api/orders/" + num)
	if err != nil {
		uc.logger.Debug("workOrder error", zap.Error(err))
		return
	}
	if resp.StatusCode == http.StatusOK {
		b, err := io.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err != nil {
			return
		}

		var acc domain.Accrual
		if err = json.Unmarshal(b, &acc); err != nil {
			uc.logger.Debug("workOrder unmarshal error", zap.Error(err))
			return
		}
		err = uc.storage.UpdateOrder(ctx, acc)
		if err != nil {
			uc.logger.Debug("workOrder update error", zap.Error(err))
			return
		}
		/*		switch acc.Status {
				case "PROCESSED":
					uc.storage.UpdateOrder(ctx,acc)
					//uc.storage.SetAccrual(ctx, acc.Order, acc.Accrual) // change status
					//uc.storage.UpdateOrder(ctx, acc)
				case "REGISTERED": //new
					uc.storage.UpdateOrder(ctx,acc)
					//uc.storage.MvTask(ctx, acc.Order) //change status
				case "PROCESSING":
					//uc.storage.MvTask(ctx, acc.Order)
				case "INVALID":
					//uc.storage.RmTask(ctx, acc.Order)
				default:
				}*/
	} else {
		uc.logger.Debug("workWithOrder not 200", zap.String("code", resp.Status))
		//todo 429 retry after
	}
	//uc.storage.
}

func (uc *usecase) GetBalance(ctx context.Context, userID int) (float64, error) {
	return uc.storage.GetBalance(ctx, userID)
}

func (uc *usecase) GetBalanceAndWithdrawn(ctx context.Context, userID int) (balance float64, withdrawn float64, err error) {
	balance, err = uc.storage.GetBalance(ctx, userID)
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
	return uc.storage.PostWithdraw(ctx, userID, order, sum)
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
