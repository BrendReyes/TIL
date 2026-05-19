package srs

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/brendreyes/til/internal/database"
	_ "github.com/mattn/go-sqlite3"
)

func newTestDB(t *testing.T) *database.Queries {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open in-memory db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	schema := `
	CREATE TABLE IF NOT EXISTS entries (
		id                   INTEGER PRIMARY KEY NOT NULL,
		body                 TEXT    NOT NULL,
		tag                  TEXT    NOT NULL,
		created_at           DATETIME NOT NULL,
		last_reviewed_at     DATETIME NOT NULL,
		review_interval_days INTEGER  NOT NULL DEFAULT 1,
		review_count         INTEGER  NOT NULL DEFAULT 0,
		ease_factor          REAL     NOT NULL DEFAULT 2.5,
		updated_at           DATETIME NOT NULL
	);`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("create schema: %v", err)
	}

	return database.New(db)
}

// seedEntry inserts one entry and returns its assigned ID.
func seedEntry(t *testing.T, q *database.Queries, body, tag string) int64 {
	t.Helper()
	now := time.Now().UTC()
	entry, err := q.CreateEntry(context.Background(), database.CreateEntryParams{
		Body:           body,
		Tag:            tag,
		CreatedAt:      now,
		UpdatedAt:      now,
		LastReviewedAt: now,
	})
	if err != nil {
		t.Fatalf("seedEntry: %v", err)
	}
	return entry.ID
}