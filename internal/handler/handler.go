package handler

import "go.uber.org/zap"

type storage interface {
}

type Handler struct {
	Storage storage
	logger  *zap.Logger
}

func New(logger *zap.Logger, storage storage) *Handler {
	return &Handler{
		logger:  logger,
		Storage: storage,
	}
}
