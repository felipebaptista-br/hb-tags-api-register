package httpapi

import (
	"archive/zip"
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/felipebaptista-br/hb-tags-api-register/internal/store"
)

func TestTagQRZip(t *testing.T) {
	pool := openTestPool(t)
	ctx := context.Background()
	handler := testHandler(store.New(pool))

	denied := perform(handler, http.MethodPost, "/api/admin/tags/qrs", `{"quantity":1}`, "")
	if denied.Code != http.StatusUnauthorized {
		t.Fatalf("status %d", denied.Code)
	}

	created := perform(handler, http.MethodPost, "/api/admin/tags/qrs", `{"quantity":2,"batch":"lote-qrs"}`, "test-admin-key")
	if created.Code != http.StatusOK {
		t.Fatalf("status %d body %s", created.Code, created.Body.String())
	}
	if created.Header().Get("Content-Type") != "application/zip" {
		t.Fatalf("content type %s", created.Header().Get("Content-Type"))
	}
	if created.Header().Get("Content-Disposition") != `attachment; filename="tags-lote-qrs.zip"` {
		t.Fatalf("disposition %s", created.Header().Get("Content-Disposition"))
	}

	files := zipJPEGs(t, created.Body.Bytes())
	if len(files) != 2 {
		t.Fatalf("files %d", len(files))
	}
	for name := range files {
		id := strings.TrimSuffix(name, ".jpg")
		var status string
		if err := pool.QueryRow(ctx, `SELECT status::text FROM tag WHERE id = $1 AND batch = 'lote-qrs'`, id).Scan(&status); err != nil {
			t.Fatal(err)
		}
		if status != "GENERATED" {
			t.Fatalf("status %s", status)
		}
	}

	var userID, categoryID string
	if err := pool.QueryRow(ctx, `
		INSERT INTO "user" (name, email) VALUES ('Ana', 'ana-qrs@example.com') RETURNING id::text
	`).Scan(&userID); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `
		INSERT INTO category (slug, name) VALUES ('pet', 'Pets') RETURNING id::text
	`).Scan(&categoryID); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO tag (id, status, batch, user_id, category_id, profile_data, activated_at)
		VALUES ('23456789AH', 'ACTIVATED', 'lote-qrs', $1, $2, '{}'::jsonb, now())
	`, userID, categoryID); err != nil {
		t.Fatal(err)
	}

	exported := perform(handler, http.MethodGet, "/api/admin/tags/qrs?batch=lote-qrs", "", "test-admin-key")
	if exported.Code != http.StatusOK {
		t.Fatalf("status %d body %s", exported.Code, exported.Body.String())
	}
	exportedFiles := zipJPEGs(t, exported.Body.Bytes())
	if _, ok := exportedFiles["23456789AH.jpg"]; ok {
		t.Fatal("activated tag was exported")
	}
	if len(exportedFiles) != 2 {
		t.Fatalf("exported files %d", len(exportedFiles))
	}

	empty := perform(handler, http.MethodGet, "/api/admin/tags/qrs?batch=lote-vazio", "", "test-admin-key")
	if empty.Code != http.StatusOK {
		t.Fatalf("status %d", empty.Code)
	}
	if got := zipJPEGs(t, empty.Body.Bytes()); len(got) != 0 {
		t.Fatalf("empty export files %d", len(got))
	}

	invalid := perform(handler, http.MethodGet, "/api/admin/tags/qrs?limit=0", "", "test-admin-key")
	if invalid.Code != http.StatusBadRequest {
		t.Fatalf("status %d", invalid.Code)
	}
}

func TestDecodeTagBatchQRDefault(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/admin/tags/qrs", strings.NewReader(`{}`))
	quantity, batch, ok := decodeTagBatch(httptest.NewRecorder(), req, qrExportLimit)
	if !ok || quantity != 1000 || batch != nil {
		t.Fatalf("quantity %d batch %v ok %v", quantity, batch, ok)
	}
}

func zipJPEGs(t *testing.T, body []byte) map[string]struct{} {
	t.Helper()
	reader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		t.Fatal(err)
	}
	files := map[string]struct{}{}
	for _, file := range reader.File {
		rc, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		var buf bytes.Buffer
		if _, err := buf.ReadFrom(rc); err != nil {
			rc.Close()
			t.Fatal(err)
		}
		rc.Close()
		if buf.Len() < 2 || buf.Bytes()[0] != 0xFF || buf.Bytes()[1] != 0xD8 {
			t.Fatalf("%s is not a jpeg", file.Name)
		}
		files[file.Name] = struct{}{}
	}
	return files
}
