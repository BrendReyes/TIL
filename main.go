/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>

*/
package main

import (
	"github.com/brendreyes/til/cmd"
	"database/sql"
	"log"
	_ "github.com/mattn/go-sqlite3"
	"github.com/brendreyes/til/internal/database"
)

type state struct {
	db *database.Queries
}

func main() {
	cmd.Execute()

	db, err := sql.Open("sqlite3", "./til.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	queries := database.New(db)

	s := &state{
		db: queries,
	}

	_ = s
}
