package handler

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Spear5030/YAGopherMart/domain"
	"github.com/go-chi/jwtauth/v5"
	"github.com/joeljunstrom/go-luhn"
	"go.uber.org/zap"
	"io"
	"net/http"
	"time"
)

type useCase interface {
	RegisterUser(ctx context.Context, login string, password string) (int, error)
	LoginUser(ctx context.Context, login string, password string) (int, error)
	PostOrder(ctx context.Context, num string, userId int) error
	GetOrders(ctx context.Context, userID int) ([]domain.Order, error)
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

func (h *Handler) PostOrder(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if !luhn.Valid(string(b)) {
		http.Error(w, "Invalid order number", http.StatusUnprocessableEntity)
		return
	}
	userID, err := getUserID(r.Context())
	if err != nil {
		h.logger.Debug("Error with JWT token", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = h.useCase.PostOrder(r.Context(), string(b), userID)
	if err != nil {
		if errors.Is(err, domain.ErrAnotherUser) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		} else if errors.Is(err, domain.ErrAlreadyUploaded) {
			http.Error(w, err.Error(), http.StatusOK)
			return
		}
	}
	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) GetOrders(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r.Context())
	if err != nil {
		h.logger.Debug("Error with JWT token", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	orders, err := h.useCase.GetOrders(r.Context(), userID)
	if len(orders) == 0 {
		http.Error(w, "No entries", http.StatusNoContent)
		return
	}
	ordersJSON, err := json.Marshal(orders)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(ordersJSON)
}

func (h *Handler) GetBalance(w http.ResponseWriter, r *http.Request) {

	_, err := getUserID(r.Context())
	if err != nil {
		h.logger.Debug("Error with JWT token", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	//balance, err := h.useCase.GetBalance(r.Context(), userID)
	// TODO need withdrawns
}

func (h *Handler) PostWithdraw(w http.ResponseWriter, r *http.Request) {
	_, err := getUserID(r.Context())
	if err != nil {
		h.logger.Debug("Error with JWT token", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func getUserID(ctx context.Context) (int, error) {
	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return 0, err
	}
	userID, ok := claims["id"].(float64)
	if !ok {
		return 0, err
	}
	return int(userID), nil
}
