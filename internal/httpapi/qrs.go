package httpapi

import (
	"archive/zip"
	"bytes"
	"errors"
	"image/jpeg"
	"image/png"
	"net/http"
	"strconv"
	"strings"

	"github.com/felipebaptista-br/hb-tags-api-register/internal/ids"
	"github.com/felipebaptista-br/hb-tags-api-register/internal/store"
	"github.com/skip2/go-qrcode"
)

const (
	qrSize        = 512
	qrJPEGQuality = 95
	qrExportLimit = 1000
)

func (a *api) adminTagQRs(w http.ResponseWriter, r *http.Request) {
	if !authorized(r, a.cfg.AdminAPIKey) {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	switch r.Method {
	case http.MethodPost:
		a.createTagQRs(w, r)
	case http.MethodGet:
		a.exportTagQRs(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed")
	}
}

func (a *api) createTagQRs(w http.ResponseWriter, r *http.Request) {
	quantity, batch, ok := decodeTagBatch(w, r, qrExportLimit)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_body")
		return
	}
	created, err := a.store.InsertGeneratedTags(r.Context(), quantity, batch, ids.Generate)
	if err != nil {
		a.logger.Error("create_tag_qrs", "request_id", requestIDFrom(r.Context()), "err", err)
		if errors.Is(err, store.ErrIDGenerationFailed) {
			writeError(w, http.StatusInternalServerError, "id_generation_failed")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	tagIDs := make([]string, 0, len(created))
	for _, tag := range created {
		tagIDs = append(tagIDs, tag.ID)
	}
	a.writeQRZip(w, zipFilename(batch), tagIDs)
}

func (a *api) exportTagQRs(w http.ResponseWriter, r *http.Request) {
	batch, limit, offset, ok := parseQRExportQuery(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_query")
		return
	}
	rows, err := a.store.ListGeneratedTags(r.Context(), batch, limit, offset)
	if err != nil {
		a.logger.Error("export_tag_qrs", "request_id", requestIDFrom(r.Context()), "err", err)
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	tagIDs := make([]string, 0, len(rows))
	for _, row := range rows {
		tagIDs = append(tagIDs, row.ID)
	}
	a.writeQRZip(w, zipFilename(batch), tagIDs)
}

func (a *api) writeQRZip(w http.ResponseWriter, filename string, ids []string) {
	files := make([]qrFile, 0, len(ids))
	for _, id := range ids {
		jpg, err := encodeQRJPEG(a.tagURL(id))
		if err != nil {
			a.logger.Error("qr_jpeg", "err", err)
			writeError(w, http.StatusInternalServerError, "internal")
			return
		}
		files = append(files, qrFile{name: id + ".jpg", data: jpg})
	}
	var body bytes.Buffer
	zw := zip.NewWriter(&body)
	for _, file := range files {
		entry, err := zw.Create(file.name)
		if err != nil {
			a.logger.Error("qr_zip", "err", err)
			writeError(w, http.StatusInternalServerError, "internal")
			return
		}
		if _, err := entry.Write(file.data); err != nil {
			a.logger.Error("qr_zip", "err", err)
			writeError(w, http.StatusInternalServerError, "internal")
			return
		}
	}
	if err := zw.Close(); err != nil {
		a.logger.Error("qr_zip", "err", err)
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body.Bytes())
}

type qrFile struct {
	name string
	data []byte
}

func encodeQRJPEG(content string) ([]byte, error) {
	pngBytes, err := qrcode.Encode(content, qrcode.Medium, qrSize)
	if err != nil {
		return nil, err
	}
	img, err := png.Decode(bytes.NewReader(pngBytes))
	if err != nil {
		return nil, err
	}
	var jpg bytes.Buffer
	if err := jpeg.Encode(&jpg, img, &jpeg.Options{Quality: qrJPEGQuality}); err != nil {
		return nil, err
	}
	return jpg.Bytes(), nil
}

func parseQRExportQuery(r *http.Request) (batch *string, limit, offset int, ok bool) {
	query := r.URL.Query()
	limit = qrExportLimit
	if raw := query.Get("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 || n > qrExportLimit {
			return nil, 0, 0, false
		}
		limit = n
	}
	if raw := query.Get("offset"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 0 {
			return nil, 0, 0, false
		}
		offset = n
	}
	if query.Has("batch") {
		value := query.Get("batch")
		if value != "" {
			if len(value) > 50 {
				return nil, 0, 0, false
			}
			batch = &value
		}
	}
	return batch, limit, offset, true
}

func zipFilename(batch *string) string {
	if batch == nil || *batch == "" {
		return "tags-qrs.zip"
	}
	safe := strings.Map(func(r rune) rune {
		switch r {
		case '"', '\\', '/', '\n', '\r':
			return '_'
		default:
			return r
		}
	}, *batch)
	return "tags-" + safe + ".zip"
}
