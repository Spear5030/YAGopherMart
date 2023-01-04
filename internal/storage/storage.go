package storage

import (
	"context"
	"database/sql"
	"errors"
	"github.com/Spear5030/YAGopherMart/domain"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
	"log"
	"time"
)

type Pinger interface {
	Ping() error
}

type storage struct {
	db     *sql.DB
	logger *zap.Logger
}

func New(logger *zap.Logger, dsn string) (*storage, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	pgs := storage{
		db:     db,
		logger: logger,
	}

	return &pgs, nil
}

func (pgs *storage) Ping() error {
	err := pgs.db.Ping()
	if err != nil {
		panic(err)
	}
	return err
}

func (pgs *storage) RegisterUser(ctx context.Context, login string, hash string) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var id int
	query := `INSERT INTO users(login, hash) 
          			VALUES($1, $2) returning id;`
	row := pgs.db.QueryRowContext(ctx, query, login, hash)
	err := row.Scan(&id)

	var pgErr *pgconn.PgError
	if err != nil {
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return 0, domain.ErrUserExists
		}
		return 0, err
	} else {
		return id, nil
	}
	return 0, err
}

func (pgs *storage) GetUserHash(ctx context.Context, login string) (id int, hash string, err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	query := `select id, hash from users 
       			where login=$1;`
	row := pgs.db.QueryRowContext(ctx, query, login)
	err = row.Scan(&id, &hash)
	if err != nil {
		return 0, "", domain.ErrInvalidPassword
	}
	return id, hash, nil
}

func (pgs *storage) PostOrder(ctx context.Context, num string, userID int) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	query := `insert into orders(number, user_id, status, uploaded_at)  
       			values($1,$2,$3,$4);`
	_, err := pgs.db.ExecContext(ctx, query, num, userID, 1, time.Now().UTC()) //1 for NEW
	var pgErr *pgconn.PgError
	if err != nil {
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			pgs.logger.Debug("Unique error")
			var userQuery int
			query = `select user_id from orders
						where number=$1`
			pgs.db.QueryRowContext(ctx, query, num).Scan(&userQuery)
			if userQuery == userID {
				return domain.ErrAlreadyUploaded
			}
			return domain.ErrAnotherUser
		}
		pgs.logger.Debug("post error", zap.Error(err))
		return err
	} else {
		return nil
	}
}

func (pgs *storage) GetOrders(ctx context.Context, userID int) ([]domain.Order, error) {
	var orders []domain.Order
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	query := `select o.number, s.status, o.uploaded_at, o.accrual from orders o
                join order_status s 
                    on	o.status=s.id 
       			where o.user_id=$1
					order by o.uploaded_at;`
	rows, err := pgs.db.QueryContext(ctx, query, userID)
	defer rows.Close()
	if err != nil {
		pgs.logger.Debug(err.Error())
		return nil, err
	}
	for rows.Next() {
		var order domain.Order
		var acc sql.NullInt64
		err = rows.Scan(&order.Number, &order.Status, &order.UploadedAt, &acc)
		if acc.Valid {
			order.Accrual = acc.Int64
		}
		if err != nil {
			pgs.logger.Debug(err.Error())
		} else {
			orders = append(orders, order)
		}
	}
	return orders, nil
}
