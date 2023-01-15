package app

import (
	"context"
	"fmt"
	"github.com/Spear5030/YAGopherMart/db/migrate"
	"github.com/Spear5030/YAGopherMart/internal/config"
	"github.com/Spear5030/YAGopherMart/internal/handler"
	"github.com/Spear5030/YAGopherMart/internal/router"
	"github.com/Spear5030/YAGopherMart/internal/storage"
	"github.com/Spear5030/YAGopherMart/internal/usecase"
	"github.com/Spear5030/YAGopherMart/pkg/logger"
	"go.uber.org/zap"
	"log"
	"math/rand"
	"net/http"
	"time"
)

type App struct {
	HTTPServer *http.Server
	logger     *zap.Logger
	ticker     *time.Ticker
}

func New(cfg config.Config) (*App, error) {
	lg, err := logger.New(true)
	if err != nil {
		lg.Debug(err.Error())
		return nil, err
	}
	repo, err := storage.New(lg, cfg.Database)
	if err != nil {
		lg.Debug(err.Error())
		return nil, err
	}
	err = migrate.Migrate(cfg.Database, migrate.Migrations)
	if err != nil {
		lg.Debug(err.Error())
		return nil, err
	}

	useCase := usecase.New(lg, repo, cfg.Accrual)
	h := handler.New(lg, useCase)
	r := router.New(h)

	t := time.NewTicker(1 * time.Second * 3)
	n := 5
	workersCount := 5
	c := make(chan string, n)
	quit := make(chan bool)
	ctx := context.Background()
	ctxWithCancel, cancel := context.WithCancel(ctx)
	go func() {
		for {
			select {
			case <-ctxWithCancel.Done():
				cancel()
			case <-t.C:
				orders, err := repo.GetOrdersForUpdate(ctx, n)
				if err != nil {
					lg.Debug(err.Error())
				}
				for _, order := range orders {
					c <- order
				}
				fmt.Println(orders)
				//lg.Debug("timer", zap.Array("orders", ))
			case <-quit:
				return
			}
		}
	}()
	for i := 0; i < workersCount; i++ {
		go func() {
			for order := range c {
				lg.Debug("Worker starts work with order ", zap.String("order", order))
				useCase.WorkWithOrder(ctxWithCancel, order)
			}
		}()
	}

	srv := &http.Server{
		Addr:    cfg.Addr,
		Handler: r,
	}
	return &App{
		HTTPServer: srv,
		ticker:     t,
	}, nil
}

func (app *App) Run() error {
	log.Print("app run")
	rand.Seed(time.Now().UnixNano())
	return app.HTTPServer.ListenAndServe()
}
