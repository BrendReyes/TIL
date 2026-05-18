package srs

import (
	"context"
	"fmt"
	"time"
	"github.com/brendreyes/til/internal/database"
)

func (s *State) AddEntry(entry string, tag string) error {
	if entry == "" || tag == "" {
		return fmt.Errorf("Body and Tag is required")
	} 

	_, err := s.DB.CreateEntry(context.Background(), database.CreateEntryParams{
		Body: entry,
		Tag:  tag,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		LastReviewedAt: time.Now().UTC(),
	})
	if err != nil {
		return fmt.Errorf("couldn't add entry: %w", err)
	}

    return nil
}