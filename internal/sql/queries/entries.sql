-- name: CreateEntry :one
INSERT INTO entries (body, tag, created_at, updated_at, last_reviewed_at) 
VALUES (
    ?,
    ?,
    ?,
    ?,
    ?
)
RETURNING *;

-- name: ListAllEntry :many
SELECT * FROM entries
ORDER BY id ASC;

-- name: ListEntryPaginated :many
SELECT * FROM entries
ORDER BY id ASC
LIMIT ? OFFSET ?;

-- name: GetEntryByID :one
SELECT * FROM entries
WHERE id = ?;

-- name: GetEntriesByTag :many
SELECT * FROM entries
WHERE LOWER(TRIM(tag)) = LOWER(TRIM(?))
ORDER BY id ASC;

-- name: CountAllEntries :one
SELECT COUNT(*) FROM entries;

-- name: DeleteEntry :execresult
DELETE FROM entries
WHERE id = ?;

-- name: DeleteAllEntries :execrows
DELETE FROM entries;

-- name: DeleteEntriesByTag :execrows
DELETE FROM entries
WHERE LOWER(TRIM(tag)) = LOWER(TRIM(?));

-- name: EditEntry :exec
UPDATE entries
SET body = ?, tag = ?, updated_at = ?
WHERE id = ?;

-- name: GetDueEntries :many
-- NOTE: last_reviewed_at is stored by the driver with a trailing timezone label
-- (e.g. "2006-01-02 15:04:05.999 +0000 UTC") that SQLite's datetime() cannot
-- parse, which makes the date arithmetic return NULL. All timestamps are written
-- as UTC, so we slice off the leading "YYYY-MM-DD HH:MM:SS" prefix that datetime()
-- can parse. See also CountDueEntries.
SELECT *
FROM entries
WHERE (
    review_count = 0
    OR datetime(substr(last_reviewed_at, 1, 19), '+' || review_interval_days || ' days') <= datetime('now')
)
ORDER BY review_count ASC, last_reviewed_at ASC;

-- name: UpdateReview :exec
UPDATE entries
SET last_reviewed_at     = ?,
    review_interval_days = ?,
    ease_factor          = ?,
    review_count         = ?,
    repetitions          = ?
WHERE id = ?;

-- name: CountEntriesByTag :many
SELECT tag, COUNT(*) AS count
FROM entries
GROUP BY tag
ORDER BY count DESC;

-- name: CountDueEntries :one
-- See GetDueEntries for why last_reviewed_at is sliced with substr().
SELECT COUNT(*) FROM entries
WHERE (
    review_count = 0
    OR datetime(substr(last_reviewed_at, 1, 19), '+' || review_interval_days || ' days') <= datetime('now')
);

-- name: CountReviewedEntries :one
SELECT COUNT(*) FROM entries
WHERE review_count > 0;

-- name: CountUnreviewedEntries :one
SELECT COUNT(*) FROM entries
WHERE review_count = 0;

-- name: ResetAllReviews :execrows
UPDATE entries
SET last_reviewed_at     = ?,
    review_interval_days = 1,
    ease_factor          = 2.5,
    review_count         = 0,
    repetitions          = 0;
