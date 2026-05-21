package srs

import "github.com/brendreyes/til/internal/database"

type State struct {
	DB *database.Queries
	DBPath string
}

