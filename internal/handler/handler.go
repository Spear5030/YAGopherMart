package handler

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Spear5030/YAGopherMart/domain"
	"go.uber.org/zap"
	"io"
	"net/http"
)

type useCase interface {
	RegisterUser(ctx context.Context, login string, password string) (string, error)
	LoginUser(ctx context.Context, login string, password string) (string, error)
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
	jwt, err := h.useCase.RegisterUser(r.Context(), userJSON.Login, userJSON.Password)

	if err != nil {
		if errors.Is(err, domain.ErrUserExists) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Authorization", "Bearer "+jwt)
}

func (h *Handler) LoginUser(w http.ResponseWriter, r *http.Request) {
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
	jwt, err := h.useCase.LoginUser(r.Context(), userJSON.Login, userJSON.Password)

	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	w.Header().Set("Authorization", "Bearer "+jwt)
}
