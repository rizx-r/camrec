package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"camrec/internal/api"
	"camrec/internal/config"
	"camrec/internal/db"
	"camrec/internal/storage"
)

type VideoHandler struct {
	store *storage.Minio
	pool  *pgxpool.Pool
	cfg   *config.Config
}

func NewVideoHandler(store *storage.Minio, pool *pgxpool.Pool, cfg *config.Config) *VideoHandler {
	return &VideoHandler{store: store, pool: pool, cfg: cfg}
}

func (h *VideoHandler) urlFor(ctx *http.Request, key string) string {
	if h.cfg.Server.PublicBucketPolicy {
		return h.store.PublicURL(key)
	}
	u, _ := h.store.PresignURL(ctx.Context(), key, time.Duration(h.cfg.Server.PresignExpireSec)*time.Second)
	return u
}

func (h *VideoHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func (h *VideoHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	vs, err := db.ListAll(r.Context(), h.pool)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	out := make([]api.VideoDTO, 0, len(vs))
	for _, v := range vs {
		out = append(out, api.VideoDTO{URL: h.urlFor(r, v.ObjectKey), Key: v.ObjectKey, StartTime: v.StartTime, EndTime: v.EndTime, SizeBytes: v.SizeBytes})
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func (h *VideoHandler) ListRange(w http.ResponseWriter, r *http.Request) {
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")
	if startStr == "" {
		http.Error(w, "start required", 400)
		return
	}
	start, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		http.Error(w, "invalid start", 400)
		return
	}
	end := time.Now()
	if endStr != "" {
		e, err := time.Parse(time.RFC3339, endStr)
		if err != nil {
			http.Error(w, "invalid end", 400)
			return
		}
		end = e
	}
	vs, err := db.ListRange(r.Context(), h.pool, start, end)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	out := make([]api.VideoDTO, 0, len(vs))
	for _, v := range vs {
		out = append(out, api.VideoDTO{URL: h.urlFor(r, v.ObjectKey), Key: v.ObjectKey, StartTime: v.StartTime, EndTime: v.EndTime, SizeBytes: v.SizeBytes})
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func (h *VideoHandler) ListLatest(w http.ResponseWriter, r *http.Request) {
	n := 10
	if q := r.URL.Query().Get("n"); q != "" {
		i, err := strconv.Atoi(q)
		if err == nil {
			n = i
		}
	}
	vs, err := db.ListLatest(r.Context(), h.pool, n)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	out := make([]api.VideoDTO, 0, len(vs))
	for _, v := range vs {
		out = append(out, api.VideoDTO{URL: h.urlFor(r, v.ObjectKey), Key: v.ObjectKey, StartTime: v.StartTime, EndTime: v.EndTime, SizeBytes: v.SizeBytes})
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}
