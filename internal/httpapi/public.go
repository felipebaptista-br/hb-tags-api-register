package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/felipebaptista-br/hb-tags-api-register/internal/ids"
	"github.com/felipebaptista-br/hb-tags-api-register/internal/visibility"
)

func (a *api) resolve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	id := r.PathValue("id")
	action := "not_found"
	if ids.Valid(id) {
		status, found, err := a.store.TagStatus(r.Context(), id)
		if err != nil {
			a.logger.Error("resolve", "request_id", requestIDFrom(r.Context()), "err", err)
			writeError(w, http.StatusInternalServerError, "internal")
			return
		}
		if found {
			action = actionForStatus(status)
		}
	}
	writeJSON(w, http.StatusOK, map[string]string{"action": action})
}

func actionForStatus(status string) string {
	switch status {
	case "GENERATED":
		return "activate"
	case "ACTIVATED":
		return "informations"
	case "LOST", "SUSPENDED":
		return "lost"
	default:
		return "not_found"
	}
}

func (a *api) publicTag(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	id := r.PathValue("id")
	if !ids.Valid(id) {
		writeError(w, http.StatusNotFound, "not_found")
		return
	}
	tag, found, err := a.store.ActivatedTag(r.Context(), id)
	if err != nil {
		a.logger.Error("public_tag", "request_id", requestIDFrom(r.Context()), "err", err)
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	if !found {
		writeError(w, http.StatusNotFound, "not_found")
		return
	}
	writeJSON(w, http.StatusOK, publicTagResponse{
		ID: tag.ID,
		Category: categoryBrief{
			Slug: tag.Slug,
			Name: tag.Name,
		},
		Profile: visibility.PublicProfile(tag.ProfileData, tag.FieldsSchema),
	})
}

type publicTagResponse struct {
	ID       string         `json:"id"`
	Category categoryBrief  `json:"category"`
	Profile  map[string]any `json:"profile"`
}

type categoryBrief struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

func (a *api) categories(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	rows, err := a.store.ListActiveCategories(r.Context())
	if err != nil {
		a.logger.Error("categories", "request_id", requestIDFrom(r.Context()), "err", err)
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	out := make([]categoryJSON, 0, len(rows))
	for _, row := range rows {
		schema := json.RawMessage(row.FieldsSchema)
		if len(schema) == 0 {
			schema = json.RawMessage("null")
		}
		out = append(out, categoryJSON{
			ID:           row.ID,
			Slug:         row.Slug,
			Name:         row.Name,
			FieldsSchema: schema,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"categories": out})
}

type categoryJSON struct {
	ID           string          `json:"id"`
	Slug         string          `json:"slug"`
	Name         string          `json:"name"`
	FieldsSchema json.RawMessage `json:"fields_schema"`
}

func (a *api) healthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := a.store.Ping(ctx); err != nil {
		a.logger.Error("healthz", "request_id", requestIDFrom(r.Context()), "err", err)
		writeError(w, http.StatusServiceUnavailable, "unavailable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
