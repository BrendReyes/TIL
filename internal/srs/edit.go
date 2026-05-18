package srs

import (
	"context"
	"fmt"
	"time"
	"strings"
	"github.com/brendreyes/til/internal/database"
)


func (s *State) EditEntry(id int64) error {
	entry, err := s.DB.GetEntry(context.Background(), id)
	if err != nil {
		fmt.Printf("Entry [#%d] Does not exist.\n", id)
		return nil
	}

	// Calling the TUI Editor goes here
	newBody, newTag, saved, err := RunEditor(entry.Body, entry.Tag)
	if err != nil {
	    return err
	}
	if !saved {
		fmt.Println("Aborted.")
		return nil
	}

	newBody = strings.TrimSpace(newBody)
	newTag = strings.TrimSpace(newTag)
	if newBody == strings.TrimSpace(entry.Body) && newTag == entry.Tag {
	    fmt.Println("No changes detected...")
	    return nil
	}

	if newBody == "" || newTag == "" {
		return fmt.Errorf("Body and Tag is required")
	}

	err = s.DB.EditEntry(context.Background(), database.EditEntryParams{
		ID:   id,
		Body: newBody,
		Tag:  newTag,
		UpdatedAt: time.Now().UTC(),
	})

	if err != nil {
		return fmt.Errorf("failed to save changes: %w", err)
	}

	fmt.Printf("✓ Updated entry #%d\n", id)
	return nil

}