package srs

import (
	"testing"
	"time"

	"github.com/brendreyes/til/internal/database"
)

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

	baseEntry := database.Entry{
		ID:                 1,
		ReviewIntervalDays: 6,
		EaseFactor:         2.5,
		ReviewCount:        2,
		LastReviewedAt:     time.Now().UTC(),
	}

	tests := []struct {
		name              string
		args              args
		wantInterval      int64
		wantEaseFactor    float64
		wantReviewCount   int64
	}{
		{
			// quality < 3 → reset: interval=1, reviewCount=0, easeFactor unchanged
			name: "Again (quality=1) resets interval and count",
			args: args{
				entry:   baseEntry,
				quality: 1,
			},
			wantInterval:    1,
			wantReviewCount: 0,
			// easeFactor is not modified on failure path in calculateNextReview
			wantEaseFactor: 2.5,
		},
		{
			// First review (reviewCount=0), Good (quality=4) → interval=1, count=1
			name: "Good on first review sets interval to 1",
			args: args{
				entry: database.Entry{
					ID:                 2,
					ReviewIntervalDays: 1,
					EaseFactor:         2.5,
					ReviewCount:        0,
					LastReviewedAt:     time.Now().UTC(),
				},
				quality: 4,
			},
			wantInterval:    1,
			wantReviewCount: 1,
			wantEaseFactor: 2.5,
		},
		{
			name: "Good on second review sets interval to 6",
			args: args{
				entry: database.Entry{
					ID:                 3,
					ReviewIntervalDays: 1,
					EaseFactor:         2.5,
					ReviewCount:        1,
					LastReviewedAt:     time.Now().UTC(),
				},
				quality: 4,
			},
			wantInterval:    6,
			wantReviewCount: 2,
			wantEaseFactor:  2.5,
		},
		{
			name: "Good on subsequent review multiplies interval by ease factor",
			args: args{
				entry:   baseEntry, // reviewCount=2, interval=6, ef=2.5
				quality: 4,
			},
			wantInterval:    15,
			wantReviewCount: 3,
			wantEaseFactor:  2.5,
		},
		{
			name: "Easy (quality=5) increases ease factor and interval",
			args: args{
				entry:   baseEntry,
				quality: 5,
			},
			wantInterval:    16,
			wantReviewCount: 3,
			wantEaseFactor:  2.6,
		},
		{
			name: "Hard (quality=3) decreases ease factor",
			args: args{
				entry:   baseEntry,
				quality: 3,
			},
			wantInterval:    14,
			wantReviewCount: 3,
			wantEaseFactor:  2.36,
		},
		{
			name: "ease factor is clamped to 1.3 minimum",
			args: args{
				entry: database.Entry{
					ID:                 5,
					ReviewIntervalDays: 6,
					EaseFactor:         1.3,
					ReviewCount:        2,
					LastReviewedAt:     time.Now().UTC(),
				},
				quality: 3,
			},
			wantInterval:    8,
			wantReviewCount: 3,
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
					LastReviewedAt:     time.Now().UTC(),
				},
				quality: 4,
			},
			wantInterval:    15,
			wantReviewCount: 3,
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
