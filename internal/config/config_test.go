package config

import "testing"

func TestLoad(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/hbtags")
	t.Setenv("ADMIN_API_KEY", "secret")
	t.Setenv("APP_URL", "https://app.example/")
	t.Setenv("FRONTEND_ORIGIN", "https://front.example")
	t.Setenv("PORT", "8080")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AppURL != "https://app.example" {
		t.Fatalf("APP_URL = %s", cfg.AppURL)
	}
	if cfg.Port != "8080" {
		t.Fatalf("PORT = %s", cfg.Port)
	}
}

func TestLoadRejectsMissingAndInvalidValues(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/hbtags")
	t.Setenv("ADMIN_API_KEY", "secret")
	t.Setenv("APP_URL", "https://app.example")
	t.Setenv("FRONTEND_ORIGIN", "https://front.example/app")
	t.Setenv("PORT", "8080")
	if _, err := Load(); err == nil {
		t.Fatal("expected origin with a path to be rejected")
	}

	t.Setenv("FRONTEND_ORIGIN", "https://front.example")
	t.Setenv("PORT", "0")
	if _, err := Load(); err == nil {
		t.Fatal("expected port 0 to be rejected")
	}

	t.Setenv("PORT", "8080")
	t.Setenv("ADMIN_API_KEY", "")
	if _, err := Load(); err == nil {
		t.Fatal("expected a missing API key to be rejected")
	}
}
