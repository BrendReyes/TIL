package main

import (
    "database/sql"
    "embed"
    "log"
    "os"
    "path/filepath"

    "github.com/brendreyes/til/cmd"
    "github.com/brendreyes/til/internal/database"
    "github.com/brendreyes/til/internal/srs"
    "github.com/pressly/goose/v3"
    _ "modernc.org/sqlite"
)

//go:embed internal/sql/schema/*.sql
var migrations embed.FS

func main() {
    db, dbPath, err := initDB()
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    queries := database.New(db)
    s := &srs.State{
        DB:     queries,
        DBPath: dbPath,
    }
	
    cmd.SetState(s)
    cmd.Execute()
}

func initDB() (*sql.DB, string, error) {
    configDir, err := os.UserConfigDir()
    if err != nil {
        return nil, "", err
    }

    appDir := filepath.Join(configDir, "til")
    if err := os.MkdirAll(appDir, 0755); err != nil {
        return nil, "", err
    }

    dbPath := filepath.Join(appDir, "til.db")
    db, err := sql.Open("sqlite", dbPath)
    if err != nil {
        return nil, "", err
    }

    goose.SetBaseFS(migrations)
    goose.SetLogger(goose.NopLogger())
    if err := goose.SetDialect("sqlite3"); err != nil {
        return nil, "", err
    }

    if err := goose.Up(db, "internal/sql/schema"); err != nil {
        return nil, "", err
    }

    return db, dbPath, nil
}