package httpapi

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type contextKey int

const requestIDKey contextKey = 1

type statusWriter struct {
	http.ResponseWriter
	status int
	wrote  bool
}

func (w *statusWriter) WriteHeader(code int) {
	if w.wrote {
		return
	}
	w.status = code
	w.wrote = true
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if !w.wrote {
		w.status = http.StatusOK
		w.wrote = true
	}
	return w.ResponseWriter.Write(b)
}

func (w *statusWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func (a *api) middleware(next http.Handler) http.Handler {
	next = a.rateLimit(next)
	next = a.timeout(next)
	next = a.logRequests(next)
	next = a.requestID(next)
	next = a.recover(next)
	return next
}

func (a *api) recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			rec := recover()
			if rec == nil {
				return
			}
			a.logger.Error("panic",
				"request_id", w.Header().Get("X-Request-ID"),
				"panic", fmt.Sprint(rec),
			)
			writeError(w, http.StatusInternalServerError, "internal")
		}()
		next.ServeHTTP(w, r)
	})
}

func (a *api) requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := acceptedRequestID(r.Header.Get("X-Request-ID"))
		w.Header().Set("X-Request-ID", id)
		ctx := context.WithValue(r.Context(), requestIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *api) logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w}
		defer func() {
			status := sw.status
			if status == 0 {
				status = http.StatusInternalServerError
			}
			a.logger.Info("request",
				"request_id", requestIDFrom(r.Context()),
				"method", r.Method,
				"path", r.URL.Path,
				"status", status,
				"duration_ms", time.Since(start).Milliseconds(),
				"remote_ip", clientIP(r),
			)
		}()
		next.ServeHTTP(sw, r)
	})
}

func (a *api) timeout(next http.Handler) http.Handler {
	body := "{\"error\":\"timeout\"}\n"
	timed := http.TimeoutHandler(next, a.requestTimeout, body)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timed.ServeHTTP(timeoutResponseWriter{ResponseWriter: w}, r)
	})
}

// timeoutResponseWriter labels the timeout response as JSON without forcing
// that content type onto CSV or empty responses.
type timeoutResponseWriter struct {
	http.ResponseWriter
}

func (w timeoutResponseWriter) WriteHeader(code int) {
	if code == http.StatusServiceUnavailable && w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "application/json")
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w timeoutResponseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func (a *api) rateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" {
			next.ServeHTTP(w, r)
			return
		}
		if !a.limiter.allow(clientIP(r)) {
			writeError(w, http.StatusTooManyRequests, "rate_limited")
			return
		}
		next.ServeHTTP(w, r)
	})
}

type ipLimiter struct {
	mu    sync.Mutex
	limit rate.Limit
	burst int
	ips   map[string]*rate.Limiter
}

func (l *ipLimiter) allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	lim, ok := l.ips[ip]
	if !ok {
		lim = rate.NewLimiter(l.limit, l.burst)
		l.ips[ip] = lim
	}
	return lim.Allow()
}

func clientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		first := strings.TrimSpace(strings.Split(forwarded, ",")[0])
		if first != "" {
			return first
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func acceptedRequestID(value string) string {
	if value == "" || len(value) > 128 {
		return newRequestID()
	}
	for _, c := range value {
		switch {
		case c >= 'a' && c <= 'z':
		case c >= 'A' && c <= 'Z':
		case c >= '0' && c <= '9':
		case c == '-' || c == '_':
		default:
			return newRequestID()
		}
	}
	return value
}

func newRequestID() string {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 36)
	}
	return hex.EncodeToString(buf[:])
}

func requestIDFrom(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	id, _ := ctx.Value(requestIDKey).(string)
	return id
}

func authorized(r *http.Request, key string) bool {
	got := sha256.Sum256([]byte(r.Header.Get("X-Api-Key")))
	want := sha256.Sum256([]byte(key))
	return subtle.ConstantTimeCompare(got[:], want[:]) == 1
}
