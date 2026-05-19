-- +goose Up
ALTER TABLE entries ADD COLUMN updated_at DATETIME NOT NULL;

-- +goose Down
ALTER TABLE entries DROP COLUMN updated_at;