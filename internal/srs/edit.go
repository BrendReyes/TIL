package srs

import (
	"context"
	"fmt"
	"database/sql"
	"github.com/brendreyes/til/internal/database"
)

func (s *State) EditEntry(id int64) error {
	entry, err := s.DB.GetEntry(context.Background(), id)
	if err != nil {
		return fmt.Errorf("[#%d] Does not exist: %w", id, err)
	}

	fmt.Printf("This will be edited:\nid: %d\nbody: %s\ntag: %s\ncreated_at: %s\n", entry.ID, entry.Body, entry.Tag, entry.CreatedAt.Format("2006-01-02 15:04:05"))

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

	err = s.DB.EditEntry(context.Background(), database.EditEntryParams{
		ID:   id,
		Body: newBody,
		Tag:  sql.NullString{
			String: newTag, 
			Valid: newTag != "",
		},
	})

	if err != nil {
		return fmt.Errorf("failed to save changes: %w", err)
	}

	fmt.Printf("✓ Updated entry #%d\n", id)
	return nil


	// below here will probably be applied in the TUI part
	/*
	_, err := s.DB.EditEntry(context.Background(), database.EditEntryParams{
		Body: body,
		Tag:  tag,
		ID: id,
	})
	if err != nil {
		return fmt.Errorf("couldn't edit entry: %w", err)
	}
	*/

    return nil
}