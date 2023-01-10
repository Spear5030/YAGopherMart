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

func (pgs *storage) GetBalance(ctx context.Context, userID int) (balance float64, err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `select balance from users 
       			where id = $1;`
	err = pgs.db.QueryRowContext(ctx, query, userID).Scan(&balance)
	if err != nil {
		pgs.logger.Debug("get balance error", zap.Error(err))
		return 0, err
	}
	return balance, nil
}

func (pgs *storage) GetWithdrawn(ctx context.Context, userID int) (withdrawn float64, err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `select sum from withdrawals 
       			where user_id = $1;`
	err = pgs.db.QueryRowContext(ctx, query, userID).Scan(&withdrawn)
	if err != nil {
		pgs.logger.Debug("get withdrawn error", zap.Error(err))
		return 0, err
	}
	return withdrawn, nil
}

func (pgs *storage) PostWithdraw(ctx context.Context, userID int, order string, sum float64) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `insert into withdrawals(number,user_id, sum)  
       			values($1,$2,$3);`
	_, err := pgs.db.ExecContext(ctx, query, order, userID, sum)
	return err
}

func (pgs *storage) UpdateOrder(ctx context.Context, accrual domain.Accrual) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if accrual.Accrual > 0 {
		// todo TX
		tx, err := pgs.db.Begin()
		if err != nil {
			pgs.logger.Debug("update TX error" + err.Error())
			return err
		}
		defer tx.Rollback()

		query := `update orders set status = s.id, accrual = $2, updated_at = $3
              		from order_status s 
				where number=$4 and s.status=$1;`
		_, err = tx.ExecContext(ctx, query, accrual.Status, accrual.Accrual, time.Now().UTC(), accrual.Order)
		if err != nil {
			pgs.logger.Debug(err.Error())
			return err
		}
		// some strange with balance. check default 0
		query = `update users set  balance = balance + $1
             		from orders o 
				where id= o.user_id and o.number=$2;`
		_, err = tx.ExecContext(ctx, query, accrual.Accrual, accrual.Order)
		if err != nil {
			pgs.logger.Debug(err.Error())
			return err
		}
		err = tx.Commit()
		if err != nil {
			pgs.logger.Debug("update TX error" + err.Error())
			return err
		}
	} else {
		query := `update orders set status = s.id, updated_at = $3
              		from order_status s 
				where number=$4 and s.status=$1;`
		_, err := pgs.db.ExecContext(ctx, query, accrual.Status, time.Now().UTC())
		if err != nil {
			pgs.logger.Debug(err.Error())
			return err
		}
	}
	return nil
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
	if err != nil {
		pgs.logger.Debug(err.Error())
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var order domain.Order
		var acc sql.NullFloat64
		err = rows.Scan(&order.Number, &order.Status, &order.UploadedAt, &acc)
		if acc.Valid {
			order.Accrual = acc.Float64
		}
		if err != nil {
			pgs.logger.Debug(err.Error())
		} else {
			orders = append(orders, order)
		}
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (pgs *storage) GetOrdersForUpdate(ctx context.Context, n int) ([]string, error) {
	var orders []string
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	query := `select o.number from orders o
                join order_status s 
                    on	o.status=s.id 
       			where s.status='NEW' or s.status='PROCESSING'
					order by o.uploaded_at
						limit $1;`
	rows, err := pgs.db.QueryContext(ctx, query, n)
	if err != nil {
		pgs.logger.Debug(err.Error())
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var order string
		err = rows.Scan(&order)
		if err != nil {
			pgs.logger.Debug(err.Error())
		} else {
			orders = append(orders, order)
		}
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func (pgs *storage) GetWithdrawals(ctx context.Context, userID int) ([]domain.Withdraw, error) {
	var withdrawals []domain.Withdraw
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	query := `select number, sum, proccesed_at from withdrawals  
       			where user_id=$1
					order by proccesed_at;`
	rows, err := pgs.db.QueryContext(ctx, query, userID)
	if err != nil {
		pgs.logger.Debug(err.Error())
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var withdrawal domain.Withdraw
		err = rows.Scan(&withdrawal.Order, &withdrawal.Sum, &withdrawal.ProcessedAt)
		if err != nil {
			pgs.logger.Debug(err.Error())
		} else {
			withdrawals = append(withdrawals, withdrawal)
		}
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return withdrawals, nil
}

func (pgs *storage) DropGoose() error {
	_, err := pgs.db.Exec("drop table goose_db_version;")
	return err
}
