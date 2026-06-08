package srs

import (
	"context"
	"testing"
	"time"

	"github.com/brendreyes/til/internal/database"
)

// TestGetDueEntries_Scheduling guards against the timestamp-serialization bug
// where a previously-reviewed but overdue entry never resurfaced because
// SQLite's datetime() could not parse the stored last_reviewed_at. It must run
// against the production driver (see helper_test.go) to be meaningful.
func TestGetDueEntries_Scheduling(t *testing.T) {
	q := newTestDB(t)
	ctx := context.Background()
	now := time.Now().UTC()

	// Reviewed 10 days ago with a 1-day interval -> overdue, must be due.
	overdue, err := q.CreateEntry(ctx, database.CreateEntryParams{
		Body: "overdue", Tag: "t", CreatedAt: now, UpdatedAt: now,
		LastReviewedAt: now.AddDate(0, 0, -10),
	})
	if err != nil {
		t.Fatalf("create overdue: %v", err)
	}
	if err := q.UpdateReview(ctx, database.UpdateReviewParams{
		ID: overdue.ID, LastReviewedAt: now.AddDate(0, 0, -10),
		ReviewIntervalDays: 1, EaseFactor: 2.5, ReviewCount: 1,
	}); err != nil {
		t.Fatalf("update overdue: %v", err)
	}

	// Reviewed just now with a 30-day interval -> not yet due.
	fresh, err := q.CreateEntry(ctx, database.CreateEntryParams{
		Body: "fresh", Tag: "t", CreatedAt: now, UpdatedAt: now,
		LastReviewedAt: now,
	})
	if err != nil {
		t.Fatalf("create fresh: %v", err)
	}
	if err := q.UpdateReview(ctx, database.UpdateReviewParams{
		ID: fresh.ID, LastReviewedAt: now,
		ReviewIntervalDays: 30, EaseFactor: 2.5, ReviewCount: 1,
	}); err != nil {
		t.Fatalf("update fresh: %v", err)
	}

	due, err := q.GetDueEntries(ctx)
	if err != nil {
		t.Fatalf("GetDueEntries: %v", err)
	}
	if len(due) != 1 {
		t.Fatalf("expected exactly 1 due entry, got %d", len(due))
	}
	if due[0].ID != overdue.ID {
		t.Errorf("expected overdue entry id=%d to be due, got id=%d", overdue.ID, due[0].ID)
	}

	count, err := q.CountDueEntries(ctx)
	if err != nil {
		t.Fatalf("CountDueEntries: %v", err)
	}
	if count != 1 {
		t.Errorf("CountDueEntries = %d, want 1", count)
	}
}

// TestLapse_KeepsReviewedStatus verifies that failing a card ("Again") restarts
// the SM-2 progression without making the card look never-reviewed again, which
// would corrupt the reviewed/unreviewed stats.
func TestLapse_KeepsReviewedStatus(t *testing.T) {
	q := newTestDB(t)
	ctx := context.Background()
	id := seedEntry(t, q, "lapse me", "t")

	entry, err := q.GetEntryByID(ctx, id)
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	// Pass it once (Good), then fail it (Again).
	if err := q.UpdateReview(ctx, calculateNextReview(entry, 4)); err != nil {
		t.Fatalf("good review: %v", err)
	}
	entry, _ = q.GetEntryByID(ctx, id)
	if err := q.UpdateReview(ctx, calculateNextReview(entry, 1)); err != nil {
		t.Fatalf("again review: %v", err)
	}

	entry, _ = q.GetEntryByID(ctx, id)
	if entry.Repetitions != 0 {
		t.Errorf("repetitions after lapse = %d, want 0", entry.Repetitions)
	}
	if entry.ReviewCount != 2 {
		t.Errorf("review_count after two reviews = %d, want 2", entry.ReviewCount)
	}

	reviewed, err := q.CountReviewedEntries(ctx)
	if err != nil {
		t.Fatalf("count reviewed: %v", err)
	}
	if reviewed != 1 {
		t.Errorf("reviewed count = %d, want 1 (lapsed card is still reviewed)", reviewed)
	}
	unreviewed, err := q.CountUnreviewedEntries(ctx)
	if err != nil {
		t.Fatalf("count unreviewed: %v", err)
	}
	if unreviewed != 0 {
		t.Errorf("unreviewed count = %d, want 0", unreviewed)
	}
}

func TestState_ReviewEntries(t *testing.T) {
	type fields struct {
		DB *database.Queries
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:    "no due entries returns no error",
			fields: fields{DB: func() *database.Queries {
				q := newTestDB(t)
				_, err := q.CreateEntry(t.Context(), database.CreateEntryParams{
					Body:           "defer runs LIFO",
					Tag:            "go",
					CreatedAt:      time.Now().UTC(),
					UpdatedAt:      time.Now().UTC(),
					LastReviewedAt: time.Now().UTC(),
				})
				if err != nil {
					t.Fatalf("seed reviewed entry: %v", err)
				}
				entries, _ := q.ListAllEntry(t.Context())
				if len(entries) > 0 {
					_ = q.UpdateReview(t.Context(), database.UpdateReviewParams{
						ID:                 entries[0].ID,
						LastReviewedAt:     time.Now().UTC(),
						ReviewIntervalDays: 30,
						EaseFactor:         2.5,
						ReviewCount:        1,
					})
				}
				return q
			}()},
			wantErr: false,
		},
		{
			name:    "empty table (nothing due) returns no error",
			fields:  fields{DB: newTestDB(t)},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &State{
				DB: tt.fields.DB,
			}
			if err := s.ReviewEntries(); (err != nil) != tt.wantErr {
				t.Errorf("State.ReviewEntries() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_calculateNextReview(t *testing.T) {
	type args struct {
		entry   database.Entry
		quality int
	}

	// review_count is deliberately distinct from repetitions to prove the two
	// move independently: repetitions drives the interval, review_count just
	// counts up by one every time.
	baseEntry := database.Entry{
		ID:                 1,
		ReviewIntervalDays: 6,
		EaseFactor:         2.5,
		ReviewCount:        5,
		Repetitions:        2,
		LastReviewedAt:     time.Now().UTC(),
	}

	tests := []struct {
		name            string
		args            args
		wantInterval    int64
		wantEaseFactor  float64
		wantReviewCount int64
		wantRepetitions int64
	}{
		{
			// quality < 3 → lapse: interval=1, repetitions reset to 0,
			// review_count still increments, easeFactor unchanged.
			name: "Again (quality=1) restarts repetitions but counts the review",
			args: args{
				entry:   baseEntry,
				quality: 1,
			},
			wantInterval:    1,
			wantReviewCount: 6,
			wantRepetitions: 0,
			wantEaseFactor:  2.5,
		},
		{
			// First repetition (repetitions=0), Good (quality=4) → interval=1
			name: "Good on first repetition sets interval to 1",
			args: args{
				entry: database.Entry{
					ID:                 2,
					ReviewIntervalDays: 1,
					EaseFactor:         2.5,
					ReviewCount:        0,
					Repetitions:        0,
					LastReviewedAt:     time.Now().UTC(),
				},
				quality: 4,
			},
			wantInterval:    1,
			wantReviewCount: 1,
			wantRepetitions: 1,
			wantEaseFactor:  2.5,
		},
		{
			name: "Good on second repetition sets interval to 6",
			args: args{
				entry: database.Entry{
					ID:                 3,
					ReviewIntervalDays: 1,
					EaseFactor:         2.5,
					ReviewCount:        1,
					Repetitions:        1,
					LastReviewedAt:     time.Now().UTC(),
				},
				quality: 4,
			},
			wantInterval:    6,
			wantReviewCount: 2,
			wantRepetitions: 2,
			wantEaseFactor:  2.5,
		},
		{
			name: "Good on subsequent repetition multiplies interval by ease factor",
			args: args{
				entry:   baseEntry, // repetitions=2, interval=6, ef=2.5
				quality: 4,
			},
			wantInterval:    15,
			wantReviewCount: 6,
			wantRepetitions: 3,
			wantEaseFactor:  2.5,
		},
		{
			name: "Easy (quality=5) increases ease factor and interval",
			args: args{
				entry:   baseEntry,
				quality: 5,
			},
			wantInterval:    16,
			wantReviewCount: 6,
			wantRepetitions: 3,
			wantEaseFactor:  2.6,
		},
		{
			name: "Hard (quality=3) decreases ease factor",
			args: args{
				entry:   baseEntry,
				quality: 3,
			},
			wantInterval:    14,
			wantReviewCount: 6,
			wantRepetitions: 3,
			wantEaseFactor:  2.36,
		},
		{
			name: "ease factor is clamped to 1.3 minimum",
			args: args{
				entry: database.Entry{
					ID:                 5,
					ReviewIntervalDays: 6,
					EaseFactor:         1.3,
					ReviewCount:        9,
					Repetitions:        2,
					LastReviewedAt:     time.Now().UTC(),
				},
				quality: 3,
			},
			wantInterval:    8,
			wantReviewCount: 10,
			wantRepetitions: 3,
			wantEaseFactor:  1.3,
		},
		{
			name: "corrupt ease factor below 1.3 is reset to 2.5",
			args: args{
				entry: database.Entry{
					ID:                 6,
					ReviewIntervalDays: 6,
					EaseFactor:         0,
					ReviewCount:        2,
					Repetitions:        2,
					LastReviewedAt:     time.Now().UTC(),
				},
				quality: 4,
			},
			wantInterval:    15,
			wantReviewCount: 3,
			wantRepetitions: 3,
			wantEaseFactor:  2.5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateNextReview(tt.args.entry, tt.args.quality)

			if got.ReviewIntervalDays != tt.wantInterval {
				t.Errorf("calculateNextReview() ReviewIntervalDays = %d, want %d",
					got.ReviewIntervalDays, tt.wantInterval)
			}
			if got.ReviewCount != tt.wantReviewCount {
				t.Errorf("calculateNextReview() ReviewCount = %d, want %d",
					got.ReviewCount, tt.wantReviewCount)
			}
			if got.Repetitions != tt.wantRepetitions {
				t.Errorf("calculateNextReview() Repetitions = %d, want %d",
					got.Repetitions, tt.wantRepetitions)
			}
			const epsilon = 0.001
			if diff := got.EaseFactor - tt.wantEaseFactor; diff > epsilon || diff < -epsilon {
				t.Errorf("calculateNextReview() EaseFactor = %.4f, want %.4f",
					got.EaseFactor, tt.wantEaseFactor)
			}
		})
	}
}

func Test_reviewModel_submitScore(t *testing.T) {
	db := newTestDB(t)
	entries := []database.Entry{
		{ID: 1, Body: "test 1", Tag: "tag1", ReviewIntervalDays: 1, ReviewCount: 0, EaseFactor: 2.5},
		{ID: 2, Body: "test 2", Tag: "tag2", ReviewIntervalDays: 1, ReviewCount: 0, EaseFactor: 2.5},
	}
	m := NewReviewModel(entries, db)

	// Test first submission (Easy - quality 5)
	m.selection = 3 // Easy
	m.submitScore()

	if m.currentIndex != 1 {
		t.Errorf("expected currentIndex 1, got %d", m.currentIndex)
	}
	if m.reviewedCount != 1 {
		t.Errorf("expected reviewedCount 1, got %d", m.reviewedCount)
	}
	if m.showAnswer {
		t.Error("expected showAnswer to be false after submission")
	}

	// Test second submission (Again - quality 1)
	m.selection = 0 // Again
	m.submitScore()

	if m.currentIndex != 2 {
		t.Errorf("expected currentIndex 2, got %d", m.currentIndex)
	}
	if m.reviewedCount != 2 {
		t.Errorf("expected reviewedCount 2, got %d", m.reviewedCount)
	}
	if !m.quitting {
		t.Error("expected quitting to be true after last entry")
	}
}
