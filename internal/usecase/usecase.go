package usecase

import "go.uber.org/zap"

type storage interface {
}

type usecase struct {
	logger  *zap.Logger
	storage storage
}

func New(logger *zap.Logger, storage storage) *usecase {
	return &usecase{
		logger:  logger,
		storage: storage,
	}
}
