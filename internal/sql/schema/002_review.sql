-- +goose Up
ALTER TABLE entries ADD COLUMN ease_factor REAL NOT NULL DEFAULT 2.5;

-- +goose Down
ALTER TABLE entries DROP COLUMN ease_factor;