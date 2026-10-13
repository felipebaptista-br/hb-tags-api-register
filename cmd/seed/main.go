package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		logger.Error("DATABASE_URL is required")
		os.Exit(1)
	}
	path := os.Getenv("SEED_FILE")
	if path == "" {
		path = "seed/seed_categories.sql"
	}
	body, err := os.ReadFile(path)
	if err != nil {
		logger.Error("read seed", "err", err, "path", path)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		logger.Error("database", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	conn, err := pool.Acquire(ctx)
	if err != nil {
		logger.Error("database", "err", err)
		os.Exit(1)
	}
	defer conn.Release()
	// Simple protocol so the file can contain several statements and dollar quotes.
	if _, err := conn.Conn().PgConn().Exec(ctx, string(body)).ReadAll(); err != nil {
		logger.Error("seed", "err", err)
		os.Exit(1)
	}
	logger.Info("seed complete")
}
