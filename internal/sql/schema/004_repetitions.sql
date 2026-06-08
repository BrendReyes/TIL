-- +goose Up
ALTER TABLE entries ADD COLUMN repetitions INTEGER NOT NULL DEFAULT 0;
UPDATE entries SET repetitions = review_count WHERE review_count > 0;

-- +goose Down
ALTER TABLE entries DROP COLUMN repetitions;
