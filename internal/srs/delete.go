package srs

import (
	"fmt"
	"context"
)

func (s *State) DeleteEntry(id int64) error {
	err := s.DB.DeleteEntry(context.Background(), id)
	if err != nil {
		return fmt.Errorf("couldn't delete entry: %w", err)
	}

	fmt.Printf("[%d] Delete Successfull\n", id)

    return nil
}