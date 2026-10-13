package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/felipebaptista-br/hb-tags-api-register/internal/ids"
	"github.com/felipebaptista-br/hb-tags-api-register/internal/store"
)

func TestCreateTagsUnauthorized(t *testing.T) {
	handler := testHandler(nil)
	for _, key := range []string{"", "wrong-key"} {
		rec := perform(handler, http.MethodPost, "/api/admin/tags", `{}`, key)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("key %q status = %d, want 401", key, rec.Code)
		}
		assertErrorCode(t, rec.Body.Bytes(), "unauthorized")
	}
}

func TestCreateTagsInvalidBody(t *testing.T) {
	handler := testHandler(nil)
	bodies := []string{
		``,
		`not-json`,
		`{"quantity":0}`,
		`{"quantity":1001}`,
		`{"quantity":-1}`,
		`{"batch":""}`,
		`{"batch":"` + strings.Repeat("a", 51) + `"}`,
	}
	for _, body := range bodies {
		rec := perform(handler, http.MethodPost, "/api/admin/tags", body, "test-admin-key")
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("body %q status = %d, want 400; body %s", body, rec.Code, rec.Body.Bytes())
		}
		assertErrorCode(t, rec.Body.Bytes(), "invalid_body")
	}
}

func TestCreateTagsCreated(t *testing.T) {
	pool := openTestPool(t)
	handler := testHandler(store.New(pool))

	rec := perform(handler, http.MethodPost, "/api/admin/tags", `{"quantity":3,"batch":"lote-a"}`, "test-admin-key")
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body %s", rec.Code, rec.Body.Bytes())
	}
	var created struct {
		Count int     `json:"count"`
		Batch *string `json:"batch"`
		Tags  []struct {
			ID     string `json:"id"`
			Status string `json:"status"`
			URL    string `json:"url"`
		} `json:"tags"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.Count != 3 || len(created.Tags) != 3 {
		t.Fatalf("count = %d, tags = %d", created.Count, len(created.Tags))
	}
	if created.Batch == nil || *created.Batch != "lote-a" {
		t.Fatalf("batch = %v", created.Batch)
	}
	seen := map[string]struct{}{}
	for _, tag := range created.Tags {
		if !ids.Valid(tag.ID) {
			t.Fatalf("invalid id %s", tag.ID)
		}
		if tag.Status != "GENERATED" {
			t.Fatalf("status = %s", tag.Status)
		}
		if tag.URL != "https://app.example/"+tag.ID {
			t.Fatalf("url = %s", tag.URL)
		}
		seen[tag.ID] = struct{}{}
	}
	if len(seen) != 3 {
		t.Fatalf("unique ids = %d", len(seen))
	}

	empty := perform(handler, http.MethodPost, "/api/admin/tags", `{}`, "test-admin-key")
	if empty.Code != http.StatusCreated {
		t.Fatalf("empty body status = %d, body %s", empty.Code, empty.Body.Bytes())
	}
	var one struct {
		Count int     `json:"count"`
		Batch *string `json:"batch"`
	}
	if err := json.Unmarshal(empty.Body.Bytes(), &one); err != nil {
		t.Fatal(err)
	}
	if one.Count != 1 || one.Batch != nil {
		t.Fatalf("count = %d, batch = %v", one.Count, one.Batch)
	}
}

func TestUnknownRouteIsJSON(t *testing.T) {
	rec := perform(testHandler(nil), http.MethodGet, "/missing", "", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d", rec.Code)
	}
	assertErrorCode(t, rec.Body.Bytes(), "not_found")
}

func assertErrorCode(t *testing.T, body []byte, code string) {
	t.Helper()
	var payload map[string]string
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("error json: %v; body %s", err, body)
	}
	if payload["error"] != code {
		t.Fatalf("error = %q, want %q", payload["error"], code)
	}
}
