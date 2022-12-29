package handler

import (
	"context"
	"encoding/json"
	"go.uber.org/zap"
	"io"
	"net/http"
)

type useCase interface {
	RegisterUser(ctx context.Context, login string, password string) error
}

type Handler struct {
	useCase useCase
	logger  *zap.Logger
}

func New(logger *zap.Logger, useCase useCase) *Handler {
	return &Handler{
		logger:  logger,
		useCase: useCase,
	}
}

func (h *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userJSON := struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}{}
	if err := json.Unmarshal(b, &userJSON); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	err = h.useCase.RegisterUser(context.Background(), userJSON.Login, userJSON.Password)

}
