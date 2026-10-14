package httpapi

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/felipebaptista-br/hb-tags-api-register/internal/config"
	"github.com/felipebaptista-br/hb-tags-api-register/internal/store"
	"golang.org/x/time/rate"
)

type api struct {
	cfg            config.Config
	store          *store.Store
	logger         *slog.Logger
	limiter        *ipLimiter
	requestTimeout time.Duration
}

// New builds the HTTP handler with recovery, request id, access log, timeout, and IP rate limit.
func New(cfg config.Config, st *store.Store, logger *slog.Logger, limit rate.Limit, burst int, requestTimeout time.Duration) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}
	if burst < 1 {
		burst = 1
	}
	if requestTimeout <= 0 {
		requestTimeout = 10 * time.Second
	}
	a := &api{
		cfg:    cfg,
		store:  st,
		logger: logger,
		limiter: &ipLimiter{
			limit: limit,
			burst: burst,
			ips:   map[string]*rate.Limiter{},
		},
		requestTimeout: requestTimeout,
	}
	return a.middleware(a.routes())
}

func (a *api) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/admin/tags/qrs", a.adminTagQRs)
	mux.HandleFunc("/api/admin/tags", a.adminTags)
	mux.HandleFunc("/api/tags/{id}/resolve", a.withCORS(a.resolve))
	mux.HandleFunc("/api/public/tags/{id}", a.withCORS(a.publicTag))
	mux.HandleFunc("/api/categories", a.withCORS(a.categories))
	mux.HandleFunc("/healthz", a.healthz)
	return routeHandler{mux: mux}
}

type routeHandler struct {
	mux *http.ServeMux
}

func (h routeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, pattern := h.mux.Handler(r)
	if pattern == "" {
		writeError(w, http.StatusNotFound, "not_found")
		return
	}
	h.mux.ServeHTTP(w, r)
}

func (a *api) withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Vary", "Origin")
		if r.Header.Get("Origin") == a.cfg.FrontendOrigin {
			w.Header().Set("Access-Control-Allow-Origin", a.cfg.FrontendOrigin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Request-ID")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}

func (a *api) tagURL(id string) string {
	return a.cfg.AppURL + "/" + id
}
