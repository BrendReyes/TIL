package srs

import (
	"context"
	"fmt"
	"time"
	"strings"
	"database/sql"
	"github.com/brendreyes/til/internal/database"
)

func (s *State) EditEntry(id int64) error {
	entry, err := s.DB.GetEntry(context.Background(), id)
	if err != nil {
		return fmt.Errorf("[#%d] Does not exist: %w", id, err)
	}

	// Calling the TUI Editor goes here
	currentTag := ""
	if entry.Tag.Valid {
		currentTag = entry.Tag.String
	}
	newBody, newTag, saved, err := RunEditor(entry.Body, currentTag)
	if err != nil {
	    return err
	}
	if !saved {
		fmt.Println("Aborted.")
		return nil
	}

	newBody = strings.TrimSpace(newBody)
	newTag = strings.TrimSpace(newTag)
	if newBody == strings.TrimSpace(entry.Body) && newTag == currentTag {
	    fmt.Println("No changes detected...")
	    return nil
	}

	err = s.DB.EditEntry(context.Background(), database.EditEntryParams{
		ID:   id,
		Body: newBody,
		Tag:  sql.NullString{
			String: newTag, 
			Valid: newTag != "",
		},
		UpdatedAt: time.Now(),
	})

	if err != nil {
		return fmt.Errorf("failed to save changes: %w", err)
	}

	fmt.Printf("✓ Updated entry #%d\n", id)
	return nil

}