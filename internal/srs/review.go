package srs

import (
	"math"
	"fmt"
	"context"
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

	// Initialize the TUI model with the due entries and the DB connection
	m := NewReviewModel(dueEntries, s.DB)
	p := tea.NewProgram(m)

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running review session: %w", err)
	}

	// Add a nice summary when done
	fmt.Printf("\n✓ Session complete. Reviewed %d items.\n", m.reviewedCount)
	return nil
}

// Quality scores mapping:
// 1 (Again) = 0 (Complex to remember)
// 2 (Hard)  = 3 (Hard but doable)
// 3 (Good)  = 4 (Not bad)
// 4 (Easy)  = 5 (EZ)

func calculateNextReview(entry database.Entry, quality int) database.UpdateReviewParams {
	interval := float64(entry.ReviewIntervalDays)
	easeFactor := entry.EaseFactor
	if easeFactor == 0 {
		easeFactor = 2.5 
	}
	reviewCount := entry.ReviewCount

	// 2. Calculate New Ease Factor
	// Formula: EF' = EF + (0.1 - (5 - q) * (0.08 + (5 - q) * 0.02))
	easeFactor = easeFactor + (0.1 - float64(5-quality)*(0.08+float64(5-quality)*0.02))
	if easeFactor < 1.3 {
		easeFactor = 1.3
	}

	// 3. Calculate New Interval and Repetitions
	if quality < 3 {
		reviewCount = 0
		interval = 1
	} else {
		// (Hard, Good, Easy)
		if reviewCount == 0 {
			interval = 1
		} else if reviewCount == 1 {
			interval = 6
		} else {
			interval = math.Round(interval * easeFactor)
		}
	}

	return database.UpdateReviewParams{
		ID:                 entry.ID,
		LastReviewedAt:     time.Now(),
		ReviewIntervalDays: int64(interval),
		EaseFactor:         easeFactor,
	}
}

