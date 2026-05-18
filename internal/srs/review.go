package srs

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/brendreyes/til/internal/database"
	tea "github.com/charmbracelet/bubbletea"
)

func (s *State) ReviewEntries() error {
	dueEntries, err := s.DB.GetDueEntries(context.Background())
	if err != nil {
		return fmt.Errorf("failed to fetch due entries: %w", err)
	}

	if len(dueEntries) == 0 {
		fmt.Println("✓ You're all caught up! Nothing due for review.")
		return nil
	}

	m := NewReviewModel(dueEntries, s.DB)
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("error running review session: %w", err)
	}
	final := finalModel.(*reviewModel)
	fmt.Printf("\n✓ Session complete. Reviewed %d items.\n", final.reviewedCount)
	return nil
}

// Quality scores (SM-2 scale):
// Again: 1  i completely forgot
// Hard:  3  hard
// Good:  4  somewhat ez
// Easy:  5  too EZ

func calculateNextReview(entry database.Entry, quality int) database.UpdateReviewParams {
	interval := float64(entry.ReviewIntervalDays)
	easeFactor := entry.EaseFactor
	if easeFactor < 1.3 {
		easeFactor = 2.5
	}
	reviewCount := entry.ReviewCount

	if quality < 3 {
		reviewCount = 0
		interval = 1
	} else {
		easeFactor = easeFactor + (0.1 - float64(5-quality)*(0.08+float64(5-quality)*0.02))
		if easeFactor < 1.3 {
			easeFactor = 1.3
		}

		if reviewCount == 0 {
			interval = 1
		} else if reviewCount == 1 {
			interval = 6
		} else {
			interval = math.Round(interval * easeFactor)
		}

		reviewCount++
	}

	return database.UpdateReviewParams{
		ID:                 entry.ID,
		LastReviewedAt:     time.Now().UTC(),
		ReviewIntervalDays: int64(interval),
		EaseFactor:         easeFactor,
		ReviewCount:        reviewCount,
	}
}