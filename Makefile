ifneq (,$(wildcard .env))
include .env
export
endif

.PHONY: migrate-up migrate-down seed run

MIGRATE = go run -tags postgres github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.2

migrate-up:
	$(MIGRATE) -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	$(MIGRATE) -path migrations -database "$(DATABASE_URL)" down

seed:
	go run ./cmd/seed

run:
	go run ./cmd/api
