package srs

import (
	"context"
	"fmt"
)

func (s *State) ShowStats() error {
	ctx := context.Background()

	total, err := s.DB.CountAllEntries(ctx)
	if err != nil {
		return fmt.Errorf("couldn't get total count: %w", err)
	}

	reviewed, err := s.DB.CountReviewedEntries(ctx)
	if err != nil {
		return fmt.Errorf("couldn't get reviewed count: %w", err)
	}

	unreviewed, err := s.DB.CountUnreviewedEntries(ctx)
	if err != nil {
		return fmt.Errorf("couldn't get unreviewed count: %w", err)
	}

	due, err := s.DB.CountDueEntries(ctx)
	if err != nil {
		return fmt.Errorf("couldn't get due count: %w", err)
	}

	byTag, err := s.DB.CountEntriesByTag(ctx)
	if err != nil {
		return fmt.Errorf("couldn't get entries by tag: %w", err)
	}

	fmt.Println("=== TIL Stats ===")
	fmt.Printf("  Total entries:    %d\n", total)
	fmt.Printf("  Reviewed:         %d\n", reviewed)
	fmt.Printf("  Unreviewed:       %d\n", unreviewed)
	fmt.Printf("  Due today:        %d\n", due)
	fmt.Println()
	fmt.Println("  Entries by tag:")
	for _, row := range byTag {
		fmt.Printf("    %-20s %d\n", row.Tag, row.Count)
	}

	return nil
}