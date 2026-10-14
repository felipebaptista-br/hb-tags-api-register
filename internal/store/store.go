package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrIDGenerationFailed means a batch could not be filled after the retry limit.
var ErrIDGenerationFailed = errors.New("id_generation_failed")

const idInsertAttempts = 5

// Store runs the SQL for the MVP endpoints.
type Store struct {
	pool *pgxpool.Pool
}

// New wraps a connection pool.
func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// Ping checks that the database accepts a connection.
func (s *Store) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

// GeneratedTag is a row created by the admin batch insert.
type GeneratedTag struct {
	ID     string
	Status string
}

// ListedTag is a tag row safe to return from the admin list.
type ListedTag struct {
	ID     string
	Status string
	Batch  *string
}

// ActivatedTag is an activated tag joined to its category.
type ActivatedTag struct {
	ID           string
	Slug         string
	Name         string
	FieldsSchema []byte
	ProfileData  []byte
}

// Category is an active category row.
type Category struct {
	ID           string
	Slug         string
	Name         string
	FieldsSchema []byte
}

// IDGenerator produces distinct candidate tag ids.
type IDGenerator func(n int) ([]string, error)

// InsertGeneratedTags inserts quantity rows with status GENERATED inside one transaction.
// Each attempt inserts the still-missing ids and ignores primary-key conflicts.
// After idInsertAttempts the transaction rolls back if the batch is incomplete.
func (s *Store) InsertGeneratedTags(ctx context.Context, quantity int, batch *string, generate IDGenerator) ([]GeneratedTag, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	created := make([]GeneratedTag, 0, quantity)
	for attempt := 0; attempt < idInsertAttempts && len(created) < quantity; attempt++ {
		ids, err := generate(quantity - len(created))
		if err != nil {
			return nil, err
		}
		rows, err := tx.Query(ctx, `
			INSERT INTO tag (id, batch)
			SELECT t.id, $2
			FROM unnest($1::text[]) AS t(id)
			ON CONFLICT DO NOTHING
			RETURNING id, status::text
		`, ids, batch)
		if err != nil {
			return nil, err
		}
		got, err := pgx.CollectRows(rows, pgx.RowToStructByPos[GeneratedTag])
		if err != nil {
			return nil, err
		}
		created = append(created, got...)
	}
	if len(created) < quantity {
		return nil, ErrIDGenerationFailed
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return created, nil
}

// ListTags returns one page of tags, optionally filtered by batch.
func (s *Store) ListTags(ctx context.Context, batch *string, limit, offset int) ([]ListedTag, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, status::text, batch
		FROM tag
		WHERE ($1::text IS NULL OR batch = $1)
		ORDER BY created_at, id
		LIMIT $2 OFFSET $3
	`, batch, limit, offset)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[ListedTag])
}

// ListGeneratedTags returns one page of tags that were never activated.
func (s *Store) ListGeneratedTags(ctx context.Context, batch *string, limit, offset int) ([]ListedTag, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, status::text, batch
		FROM tag
		WHERE status = 'GENERATED'
		  AND ($1::text IS NULL OR batch = $1)
		ORDER BY created_at, id
		LIMIT $2 OFFSET $3
	`, batch, limit, offset)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[ListedTag])
}

// TagStatus returns the tag status. found is false when the id does not exist.
func (s *Store) TagStatus(ctx context.Context, id string) (status string, found bool, err error) {
	err = s.pool.QueryRow(ctx, `SELECT status::text FROM tag WHERE id = $1`, id).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return status, true, nil
}

// ActivatedTag loads a tag only when its status is ACTIVATED.
func (s *Store) ActivatedTag(ctx context.Context, id string) (ActivatedTag, bool, error) {
	var tag ActivatedTag
	err := s.pool.QueryRow(ctx, `
		SELECT t.id, c.slug, c.name, c.fields_schema, t.profile_data
		FROM tag t
		JOIN category c ON c.id = t.category_id
		WHERE t.id = $1 AND t.status = 'ACTIVATED'
	`, id).Scan(&tag.ID, &tag.Slug, &tag.Name, &tag.FieldsSchema, &tag.ProfileData)
	if errors.Is(err, pgx.ErrNoRows) {
		return ActivatedTag{}, false, nil
	}
	if err != nil {
		return ActivatedTag{}, false, err
	}
	return tag, true, nil
}

// ListActiveCategories returns categories with active = true, ordered by name.
func (s *Store) ListActiveCategories(ctx context.Context) ([]Category, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, slug, name, fields_schema
		FROM category
		WHERE active
		ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[Category])
}
