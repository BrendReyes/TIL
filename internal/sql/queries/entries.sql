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
SELECT * FROM entries;

-- name: DeleteEntry :execresult
DELETE FROM entries
WHERE id = ?;

-- name: GetEntry :one
SELECT * FROM entries
WHERE id = ?;

-- name: EditEntry :exec
UPDATE entries
SET body = ?, tag = ?, updated_at = ?
WHERE id = ?;

-- name: GetDueEntries :many
SELECT *
FROM entries
WHERE review_count = 0
   OR datetime(last_reviewed_at, '+' || review_interval_days || ' days') <= datetime('now')
ORDER BY review_count ASC, last_reviewed_at ASC;

-- name: UpdateReview :exec
UPDATE entries
SET last_reviewed_at     = ?,
    review_interval_days = ?,
    ease_factor          = ?,
    review_count         = ?
WHERE id = ?;

-- name: GetEntryByID :one
SELECT * FROM entries
WHERE id = ?;
