-- +goose Up
CREATE TABLE entries (
    id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    body TEXT NOT NULL,
    tag TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_reviewed_at DATETIME DEFAULT NULL,
    review_interval_days INTEGER NOT NULL DEFAULT 1,
    review_count INTEGER NOT NULL DEFAULT 0
);

-- +goose Down
DROP TABLE entries;