package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"camrec/internal/handler"
)

func New(h *handler.VideoHandler) http.Handler {
	r := chi.NewRouter()
	r.Get("/health", h.Health)
	r.Get("/videos", h.ListAll)
	r.Get("/videos/range", h.ListRange)
	r.Get("/videos/latest", h.ListLatest)
	return r
}
