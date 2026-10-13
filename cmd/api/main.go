package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/felipebaptista-br/hb-tags-api-register/internal/config"
	"github.com/felipebaptista-br/hb-tags-api-register/internal/httpapi"
	"github.com/felipebaptista-br/hb-tags-api-register/internal/store"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/time/rate"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("config", "err", err)
		os.Exit(1)
	}

	poolCtx, poolCancel := context.WithTimeout(context.Background(), 5*time.Second)
	pool, err := pgxpool.New(poolCtx, cfg.DatabaseURL)
	if err != nil {
		poolCancel()
		logger.Error("database", "err", err)
		os.Exit(1)
	}
	if err := pool.Ping(poolCtx); err != nil {
		pool.Close()
		poolCancel()
		logger.Error("database", "err", err)
		os.Exit(1)
	}
	poolCancel()

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           httpapi.New(cfg, store.New(pool), logger, rate.Limit(5), 20, 10*time.Second),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("listening", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		pool.Close()
		logger.Error("server", "err", err)
		os.Exit(1)
	case <-stop:
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		pool.Close()
		logger.Error("shutdown", "err", err)
		os.Exit(1)
	}
	pool.Close()
	logger.Info("stopped")
}
