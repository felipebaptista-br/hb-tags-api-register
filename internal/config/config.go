package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

// Config holds the environment required to start the API.
type Config struct {
	DatabaseURL    string
	AdminAPIKey    string
	AppURL         string
	FrontendOrigin string
	Port           string
}

// Load reads and validates the process environment. It returns an error
// naming every missing variable so the process can exit before it serves.
func Load() (Config, error) {
	cfg := Config{
		DatabaseURL:    strings.TrimSpace(os.Getenv("DATABASE_URL")),
		AdminAPIKey:    os.Getenv("ADMIN_API_KEY"),
		AppURL:         strings.TrimRight(strings.TrimSpace(os.Getenv("APP_URL")), "/"),
		FrontendOrigin: strings.TrimRight(strings.TrimSpace(os.Getenv("FRONTEND_ORIGIN")), "/"),
		Port:           strings.TrimSpace(os.Getenv("PORT")),
	}

	var missing []string
	if cfg.DatabaseURL == "" {
		missing = append(missing, "DATABASE_URL")
	}
	if cfg.AdminAPIKey == "" {
		missing = append(missing, "ADMIN_API_KEY")
	}
	if cfg.AppURL == "" {
		missing = append(missing, "APP_URL")
	}
	if cfg.FrontendOrigin == "" {
		missing = append(missing, "FRONTEND_ORIGIN")
	}
	if cfg.Port == "" {
		missing = append(missing, "PORT")
	}
	if len(missing) > 0 {
		return Config{}, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	appURL, err := url.Parse(cfg.AppURL)
	if err != nil || !absoluteHTTP(appURL) {
		return Config{}, fmt.Errorf("APP_URL must be an absolute http or https URL")
	}
	origin, err := url.Parse(cfg.FrontendOrigin)
	if err != nil || !httpOrigin(origin) {
		return Config{}, fmt.Errorf("FRONTEND_ORIGIN must be an http or https origin without a path")
	}
	port, err := strconv.Atoi(cfg.Port)
	if err != nil || port < 1 || port > 65535 {
		return Config{}, fmt.Errorf("PORT must be an integer between 1 and 65535")
	}
	return cfg, nil
}

func absoluteHTTP(u *url.URL) bool {
	return u != nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

func httpOrigin(u *url.URL) bool {
	return absoluteHTTP(u) && u.Path == "" && u.RawQuery == "" && u.Fragment == ""
}
