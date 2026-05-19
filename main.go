/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>

*/
package main

import (
	"database/sql"
	"log"

	"github.com/brendreyes/til/cmd"
	"github.com/brendreyes/til/internal/database"
	"github.com/brendreyes/til/internal/srs"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "./til.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	queries := database.New(db)

	s := &srs.State{
		DB: queries,
	}

	cmd.SetState(s)
	cmd.Execute()
}
