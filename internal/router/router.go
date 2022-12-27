package router

import (
	"github.com/Spear5030/YAGopherMart/internal/handler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
)

func New(h *handler.Handler) http.Handler {
	r := chi.NewRouter()
	//r.Use(handler.CheckCookies(h.SecretKey))
	r.Use(middleware.Logger)
	r.Use(middleware.Compress(5))
	//r.Use(handler.DecompressGZRequest)
	//r.Get("/{id}", h.GetURL)
	//r.Get("/ping", h.PingDB)
	//r.Post("/", h.PostURL)
	//
	//r.Group(func(r chi.Router) {
	//	r.Use(middleware.SetHeader("Content-Type", "application/json"))
	//	r.Post("/api/shorten", h.PostJSON)
	//	r.Get("/api/user/urls", h.GetURLsByUser)
	//	r.Delete("/api/user/urls", h.DeleteBatchByUser)
	//	r.Post("/api/shorten/batch", h.PostBatch)
	//})

	return r
}
