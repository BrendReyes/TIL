package srs

import (
	"context"
	"fmt"
	"database/sql"
	"time"
	"github.com/brendreyes/til/internal/database"
)

func (s *State) AddEntry(entry string, tag string) error {
	var nullTag sql.NullString

    if tag != "" {
        nullTag = sql.NullString{
			String: tag, 
			Valid: true,
		}
    }
	
	_, err := s.DB.CreateEntry(context.Background(), database.CreateEntryParams{
		Body: entry,
		Tag:  nullTag,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		LastReviewedAt: time.Now().UTC(),
	})
	if err != nil {
		return fmt.Errorf("couldn't add entry: %w", err)
	}

    return nil
}