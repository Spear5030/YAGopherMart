package router

import (
	"github.com/Spear5030/YAGopherMart/internal/handler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	"net/http"
)

func New(h *handler.Handler) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Compress(5))
	r.Use(jwtauth.Verifier(h.JWT))
	r.Use(handler.DecompressGZRequest)
	r.Group(func(r chi.Router) {
		r.Post("/api/user/register", h.RegisterUser)
		r.Post("/api/user/login", h.LoginUser)
	})
	r.Group(func(r chi.Router) {
		r.Use(jwtauth.Authenticator)
		r.Post("/api/user/orders", h.PostOrder)
		r.Post("/api/user/balance/withdraw", h.PostWithdraw)
	})
	r.Group(func(r chi.Router) {
		r.Use(jwtauth.Authenticator)
		r.Use(middleware.SetHeader("Content-Type", "application/json"))
		r.Get("/api/user/orders", h.GetOrders)
		r.Get("/api/user/balance", h.GetBalanceAndWithdrawn)
		r.Get("/api/user/withdrawals", h.GetWithdrawals)
	})

	return r
}
