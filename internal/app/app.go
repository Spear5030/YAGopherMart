package app

import (
	"github.com/Spear5030/YAGopherMart/db/migrate"
	"github.com/Spear5030/YAGopherMart/internal/config"
	"github.com/Spear5030/YAGopherMart/internal/handler"
	"github.com/Spear5030/YAGopherMart/internal/router"
	"github.com/Spear5030/YAGopherMart/internal/storage"
	"github.com/Spear5030/YAGopherMart/internal/usecase"
	"github.com/Spear5030/YAGopherMart/pkg/logger"
	"go.uber.org/zap"
	"math/rand"
	"net/http"
	"time"
)

type App struct {
	HTTPServer *http.Server
	logger     *zap.Logger
}

func New(cfg config.Config) (*App, error) {

	lg, err := logger.New(true)
	if err != nil {
		return nil, err
	}
	err = migrate.Migrate(cfg.Database, migrate.Migrations)
	if err != nil {
		lg.Debug(err.Error())
		return nil, err
	}
	repo, err := storage.New(lg, cfg.Database)
	useCase := usecase.New(lg, repo, cfg.Accrual)
	h := handler.New(lg, useCase)
	r := router.New(h)
	srv := &http.Server{
		Addr:    cfg.Addr,
		Handler: r,
	}
	return &App{
		HTTPServer: srv,
	}, nil
}

func (app *App) Run() error {
	rand.Seed(time.Now().UnixNano())
	return app.HTTPServer.ListenAndServe()
}
