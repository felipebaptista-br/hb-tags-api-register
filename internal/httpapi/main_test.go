package httpapi

import (
	"context"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/felipebaptista-br/hb-tags-api-register/internal/config"
	"github.com/felipebaptista-br/hb-tags-api-register/internal/store"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/time/rate"
)

func TestMain(m *testing.M) {
	if os.Getenv("TEST_DATABASE_URL") != "" {
		if err := ensureSchema(); err != nil {
			log.Fatal(err)
		}
	}
	os.Exit(m.Run())
}

func ensureSchema() error {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return os.ErrNotExist
	}
	path := filepath.Join(filepath.Dir(file), "..", "..", "migrations", "000001_init.up.sql")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, os.Getenv("TEST_DATABASE_URL"))
	if err != nil {
		return err
	}
	defer pool.Close()

	var exists bool
	if err := pool.QueryRow(ctx, `SELECT EXISTS (
		SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'tag'
	)`).Scan(&exists); err != nil {
		return err
	}
	if exists {
		return nil
	}
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	_, err = conn.Conn().PgConn().Exec(ctx, string(body)).ReadAll()
	return err
}

func testHandler(st *store.Store) http.Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := config.Config{
		DatabaseURL:    "postgres://test",
		AdminAPIKey:    "test-admin-key",
		AppURL:         "https://app.example",
		FrontendOrigin: "https://front.example",
		Port:           "8080",
	}
	return New(cfg, st, logger, rate.Inf, 1, 5*time.Second)
}

func openTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if _, err := pool.Exec(ctx, `TRUNCATE tag, category, "user"`); err != nil {
		t.Fatal(err)
	}
	return pool
}

func perform(handler http.Handler, method, target, body, apiKey string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	if apiKey != "" {
		req.Header.Set("x-api-key", apiKey)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}
