package handler

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Spear5030/YAGopherMart/domain"
	"github.com/go-chi/jwtauth/v5"
	"go.uber.org/zap"
	"io"
	"net/http"
	"time"
)

type useCase interface {
	RegisterUser(ctx context.Context, login string, password string) (int, error)
	LoginUser(ctx context.Context, login string, password string) (int, error)
}

type Handler struct {
	useCase useCase
	logger  *zap.Logger
	JWT     *jwtauth.JWTAuth
}

func New(logger *zap.Logger, useCase useCase) *Handler {
	tokenAuth := jwtauth.New("HS256", []byte("SecretKey"), nil) //TODO Config
	return &Handler{
		logger:  logger,
		useCase: useCase,
		JWT:     tokenAuth,
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
	id, err := h.useCase.RegisterUser(r.Context(), userJSON.Login, userJSON.Password)
	if err != nil {
		if errors.Is(err, domain.ErrUserExists) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, tokenString, err := h.JWT.Encode(map[string]interface{}{
		"id":        id,
		"ExpiredAt": time.Now().Add(time.Hour * 24),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Authorization", "Bearer "+tokenString)
	w.WriteHeader(http.StatusOK)
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
	id, err := h.useCase.LoginUser(r.Context(), userJSON.Login, userJSON.Password)

	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	_, tokenString, err := h.JWT.Encode(map[string]interface{}{
		"id":        id,
		"ExpiredAt": time.Now().Add(time.Hour * 24),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Authorization", "Bearer "+tokenString)
	w.WriteHeader(http.StatusOK)
}
