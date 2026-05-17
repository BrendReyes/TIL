package srs

import (
	"context"
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
		CreatedAt: time.Now(),
	})
	if err != nil {
		return err
	}

    return nil
}