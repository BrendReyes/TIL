package srs

import (
	"context"
	"fmt"
	"github.com/brendreyes/til/internal/database"
)

func (s *State) ListEntry() error {
	entries, err := s.DB.ListAllEntry(context.Background())
	if err != nil {
		return fmt.Errorf("couldn't list entries: %w", err)
	}

	if len(entries) == 0 {
    	fmt.Println("No entries found.")
    	return nil
	}

	fmt.Println("what you have learned:")
	for _, entry := range entries {
		printEntry(entry)
	}

    return nil
}

func printEntry(entry database.Entry) {
	
    fmt.Printf("--- Entry #%d ---\n", entry.ID)
    fmt.Printf("  Body:             %s\n", entry.Body)

    if entry.Tag.Valid {
        fmt.Printf("  Tag:              %s\n", entry.Tag.String)
    } else {
        fmt.Printf("  Tag:              (none)\n")
    }

    fmt.Printf("  Created:          %s\n", entry.CreatedAt.Format("2006-01-02 15:04:05"))

    if entry.LastReviewedAt.Valid {
        fmt.Printf("  Last Reviewed:    %s\n", entry.LastReviewedAt.Time.Format("2006-01-02 15:04:05"))
    } else {
        fmt.Printf("  Last Reviewed:    (never)\n")
    }

    //fmt.Printf("  Review Interval:  %d days\n", entry.ReviewIntervalDays)
    fmt.Printf("  Review Count:     %d\n", entry.ReviewCount)
    fmt.Println("----------------")
}