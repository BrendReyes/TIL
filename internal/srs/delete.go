package srs

import (
	"context"
	"fmt"
)

func (s *State) DeleteEntry(id int64) error {
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