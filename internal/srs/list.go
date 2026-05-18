package srs

import (
	"context"
	"fmt"
	"github.com/brendreyes/til/internal/database"
)

func (s *State) ListEntry() error {
	entries, err := s.DB.ListAllEntry(context.Background())
	if err != nil {
		return fmt.Errorf("Couldn't list entries: %w", err)
	}

	if len(entries) == 0 {
    	fmt.Println("No entries found. Use 'til add <body> -t <tag>' to start.")
    	return nil
	}

	for _, entry := range entries {
		printEntry(entry)
	}

    return nil
}

func (s *State) GetSpecificEntry(id int64) error {
	entry, err := s.DB.GetEntryByID(context.Background(), id)
	if err != nil {
		fmt.Printf("Entry [#%d] does not exist.\n", id)
		return nil
	}

	printEntry(entry)	

	return nil
}

func (s *State) ListEntriesByTag(tag string) error {
	entries, err := s.DB.GetEntriesByTag(context.Background(), tag)
	if err != nil {
		return fmt.Errorf("Couldn't fetch by tag: %w", err)
	}

	if len(entries) == 0 {
		fmt.Printf("No entries found with '%s' \n", tag)
		return nil
	}

	for _, entry := range entries {
		printEntry(entry)	
	}

	return nil
}

func printEntry(entry database.Entry) {
	
    fmt.Printf("--- Entry #%d ---\n", entry.ID)
    fmt.Printf("  Body:             %s\n", entry.Body)
    fmt.Printf("  Tag:              %s\n", entry.Tag)
    fmt.Printf("  Created:          %s\n", entry.CreatedAt.Format("2006-01-02 15:04:05"))
    fmt.Printf("  Last Reviewed:    %s\n", entry.LastReviewedAt.Format("2006-01-02 15:04:05"))
    //fmt.Printf("  Review Interval:  %d days\n", entry.ReviewIntervalDays)
    fmt.Printf("  Review Count:     %d\n", entry.ReviewCount)
    fmt.Println("----------------")
}