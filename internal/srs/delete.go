package srs

import (
	"context"
	"fmt"
)

func (s *State) DeleteEntry(id int64) error {
	fmt.Printf("Delete entry #%d? [y/N] ", id)
	var confirm string
	fmt.Scanln(&confirm)
	if confirm != "y" && confirm != "Y" {
		fmt.Println("Delete aborted.")
		return nil
	}

	result, err := s.DB.DeleteEntry(context.Background(), id)
	if err != nil {
		return fmt.Errorf("couldn't delete entry: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("couldn't verify deletion: %w", err)
	}
	if rows == 0 {
		fmt.Printf("entry [#%d] not found\n", id)
		return nil
	}

	fmt.Printf("[%d] Deleted Successfully\n", id)
	return nil
}

func (s *State) RemoveAllEntry() error {
	fmt.Printf("⚠️ WARNING, all entries will be permanently deleted. Are you sure? [y/N] ", )
	var confirm string
	fmt.Scanln(&confirm)
	if confirm != "y" && confirm != "Y" {
		fmt.Println("Delete aborted.")
		return nil
	}

    rowsDeleted, err := s.DB.DeleteAllEntries(context.Background())
    if err != nil {
        return fmt.Errorf("couldn't delete entries: %w", err)
    }

    fmt.Printf("%d entries deleted\n", rowsDeleted)
    return nil
}
