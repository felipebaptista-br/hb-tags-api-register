package httpapi

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/felipebaptista-br/hb-tags-api-register/internal/ids"
	"github.com/felipebaptista-br/hb-tags-api-register/internal/store"
)

const maxCreateBody = 4096

func (a *api) adminTags(w http.ResponseWriter, r *http.Request) {
	if !authorized(r, a.cfg.AdminAPIKey) {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	switch r.Method {
	case http.MethodPost:
		a.createTags(w, r)
	case http.MethodGet:
		a.listTags(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed")
	}
}

func (a *api) createTags(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Quantity *int    `json:"quantity"`
		Batch    *string `json:"batch"`
	}
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxCreateBody))
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body")
		return
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "invalid_body")
		return
	}

	quantity := 1
	if req.Quantity != nil {
		quantity = *req.Quantity
	}
	if quantity < 1 || quantity > 1000 {
		writeError(w, http.StatusBadRequest, "invalid_body")
		return
	}
	if req.Batch != nil && (len(*req.Batch) < 1 || len(*req.Batch) > 50) {
		writeError(w, http.StatusBadRequest, "invalid_body")
		return
	}

	created, err := a.store.InsertGeneratedTags(r.Context(), quantity, req.Batch, ids.Generate)
	if err != nil {
		if errors.Is(err, store.ErrIDGenerationFailed) {
			a.logger.Error("create_tags", "request_id", requestIDFrom(r.Context()), "err", err)
			writeError(w, http.StatusInternalServerError, "id_generation_failed")
			return
		}
		a.logger.Error("create_tags", "request_id", requestIDFrom(r.Context()), "err", err)
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}

	tags := make([]createdTagJSON, 0, len(created))
	for _, tag := range created {
		tags = append(tags, createdTagJSON{
			ID:     tag.ID,
			Status: tag.Status,
			URL:    a.tagURL(tag.ID),
		})
	}
	writeJSON(w, http.StatusCreated, createTagsResponse{
		Count: len(tags),
		Batch: req.Batch,
		Tags:  tags,
	})
}

type createTagsResponse struct {
	Count int              `json:"count"`
	Batch *string          `json:"batch"`
	Tags  []createdTagJSON `json:"tags"`
}

type createdTagJSON struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	URL    string `json:"url"`
}

func (a *api) listTags(w http.ResponseWriter, r *http.Request) {
	limit, offset, batch, format, ok := parseListQuery(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_query")
		return
	}
	rows, err := a.store.ListTags(r.Context(), batch, limit, offset)
	if err != nil {
		a.logger.Error("list_tags", "request_id", requestIDFrom(r.Context()), "err", err)
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	if format == "csv" {
		a.writeTagsCSV(w, rows)
		return
	}
	tags := make([]listedTagJSON, 0, len(rows))
	for _, row := range rows {
		tags = append(tags, listedTagJSON{
			ID:     row.ID,
			Status: row.Status,
			Batch:  row.Batch,
			URL:    a.tagURL(row.ID),
		})
	}
	writeJSON(w, http.StatusOK, listTagsResponse{
		Tags:   tags,
		Limit:  limit,
		Offset: offset,
	})
}

type listTagsResponse struct {
	Tags   []listedTagJSON `json:"tags"`
	Limit  int             `json:"limit"`
	Offset int             `json:"offset"`
}

type listedTagJSON struct {
	ID     string  `json:"id"`
	Status string  `json:"status"`
	Batch  *string `json:"batch"`
	URL    string  `json:"url"`
}

func parseListQuery(r *http.Request) (limit, offset int, batch *string, format string, ok bool) {
	query := r.URL.Query()
	limit = 50
	if raw := query.Get("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 || n > 200 {
			return 0, 0, nil, "", false
		}
		limit = n
	}
	if raw := query.Get("offset"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 0 {
			return 0, 0, nil, "", false
		}
		offset = n
	}
	if query.Has("batch") {
		value := query.Get("batch")
		if value != "" {
			if len(value) > 50 {
				return 0, 0, nil, "", false
			}
			batch = &value
		}
	}
	format = query.Get("format")
	if format != "" && format != "json" && format != "csv" {
		return 0, 0, nil, "", false
	}
	return limit, offset, batch, format, true
}

func (a *api) writeTagsCSV(w http.ResponseWriter, rows []store.ListedTag) {
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	out := csv.NewWriter(w)
	_ = out.Write([]string{"id", "status", "batch", "url"})
	for _, row := range rows {
		batch := ""
		if row.Batch != nil {
			batch = *row.Batch
		}
		_ = out.Write([]string{row.ID, row.Status, batch, a.tagURL(row.ID)})
	}
	out.Flush()
}

func methodNotAllowed(w http.ResponseWriter) {
	writeError(w, http.StatusMethodNotAllowed, "method_not_allowed")
}
