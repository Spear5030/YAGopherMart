package storage

import (
	"go.uber.org/zap"
)

type storage struct {
	logger *zap.Logger
}

func New(logger *zap.Logger) *storage {
	return &storage{
		logger: logger,
	}
}
