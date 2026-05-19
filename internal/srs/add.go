package srs

import (
	"context"
	"fmt"
	"time"
	"github.com/brendreyes/til/internal/database"
)

func (s *State) AddEntry(entry string, tag string) error {
	if entry == "" || tag == "" {
		return fmt.Errorf("Body or Tag cannot be empty.")
	}
	
	if len(entry) > 1000 {
		return fmt.Errorf("characters exceeded for the content.... (max 1000)")
	}

	if len(tag) > 100 {
		return fmt.Errorf("characters exceeded for the tag.... (max 100)")
	}

	_, err := s.DB.CreateEntry(context.Background(), database.CreateEntryParams{
		Body: entry,
		Tag:  tag,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		LastReviewedAt: time.Now().UTC(),
	})
	if err != nil {
		return fmt.Errorf("Couldn't add entry: %w", err)
	}

    return nil
}