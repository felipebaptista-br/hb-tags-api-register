package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/felipebaptista-br/hb-tags-api-register/internal/store"
)

func TestResolve(t *testing.T) {
	pool := openTestPool(t)
	ctx := context.Background()

	if _, err := pool.Exec(ctx, `INSERT INTO tag (id) VALUES ('23456789AB')`); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO tag (id, status) VALUES ('23456789AD', 'LOST')`); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO tag (id, status) VALUES ('23456789AE', 'SUSPENDED')`); err != nil {
		t.Fatal(err)
	}

	var userID, categoryID string
	if err := pool.QueryRow(ctx, `
		INSERT INTO "user" (name, email) VALUES ('Ana', 'ana@example.com') RETURNING id::text
	`).Scan(&userID); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `
		INSERT INTO category (slug, name, fields_schema)
		VALUES ('pet', 'Pets', '{"version":1,"fields":[{"key":"identity.display_name","visibility":"public"}]}'::jsonb)
		RETURNING id::text
	`).Scan(&categoryID); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO tag (id, status, user_id, category_id, profile_data, activated_at)
		VALUES ('23456789AC', 'ACTIVATED', $1, $2, '{"identity":{"display_name":"Rex"}}'::jsonb, now())
	`, userID, categoryID); err != nil {
		t.Fatal(err)
	}

	handler := testHandler(store.New(pool))
	tests := []struct {
		name   string
		id     string
		action string
	}{
		{name: "generated", id: "23456789AB", action: "activate"},
		{name: "activated", id: "23456789AC", action: "informations"},
		{name: "lost", id: "23456789AD", action: "lost"},
		{name: "suspended", id: "23456789AE", action: "lost"},
		{name: "missing", id: "23456789AF", action: "not_found"},
		{name: "malformed", id: "0000000000", action: "not_found"},
	}

	var notFoundStatus int
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := perform(handler, http.MethodGet, "/api/tags/"+tt.id+"/resolve", "", "")
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, body %s", rec.Code, rec.Body.Bytes())
			}
			if rec.Header().Get("Cache-Control") != "no-store" {
				t.Fatalf("Cache-Control = %q", rec.Header().Get("Cache-Control"))
			}
			var payload map[string]any
			if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
				t.Fatal(err)
			}
			if len(payload) != 1 {
				t.Fatalf("payload = %#v", payload)
			}
			if payload["action"] != tt.action {
				t.Fatalf("action = %v, want %s", payload["action"], tt.action)
			}
			if tt.action == "not_found" {
				if notFoundStatus == 0 {
					notFoundStatus = rec.Code
				} else if rec.Code != notFoundStatus {
					t.Fatalf("not_found status = %d, previous = %d", rec.Code, notFoundStatus)
				}
			}
		})
	}
}
