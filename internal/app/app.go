package app

import (
	"context"
	"fmt"
	"github.com/Spear5030/YAGopherMart/db/migrate"
	"github.com/Spear5030/YAGopherMart/internal/accrualer"
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
	HTTPServer    *http.Server
	logger        *zap.Logger
	ticker        *time.Ticker
	AcrrualClient *accrualer.Client
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

	acClient := accrualer.New(cfg.Accrual)

	useCase := usecase.New(lg, repo, acClient)
	h := handler.New(lg, useCase, cfg.Key)
	r := router.New(h)

	n := 100
	workersCount := 5
	c := make(chan string, n)
	ctx := context.Background()
	nop := false //boolean for nop if 429(retry)
	t := time.NewTicker(1 * time.Second * 5)
	go func() {
		for {
			select {
			case <-ctx.Done():
				lg.Debug("Context cancelled")
				return
			case <-t.C:
				nop = false
				orders, err := repo.GetOrdersForUpdate(ctx, n)
				if err != nil {
					lg.Debug(err.Error())
				}
				for _, order := range orders {
					c <- order
				}
				fmt.Println(orders)
				//lg.Debug("timer", zap.Array("orders", ))
				// case ch with sleep
			}
		}
	}()
	for i := 0; i < workersCount; i++ {
		go func() {
			for order := range c {
				if !nop {
					lg.Debug("Worker starts work with order ", zap.String("order", order))
					err = useCase.WorkWithOrder(ctx, order)
					if err != nil {
						nop = true
					}
				}
				lg.Debug("Have nop - wont work")
				// 429 error or some.
			}
		}()
	}

	srv := &http.Server{
		Addr:    cfg.Addr,
		Handler: r,
	}
	return &App{
		HTTPServer:    srv,
		ticker:        t,
		AcrrualClient: acClient,
	}, nil
}

func (app *App) Run() error {
	log.Print("app run")
	rand.Seed(time.Now().UnixNano())
	return app.HTTPServer.ListenAndServe()
}
