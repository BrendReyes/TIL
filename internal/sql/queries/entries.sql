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
ORDER BY created_at DESC;;

-- name: ListEntryPaginated :many
SELECT * FROM entries
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: GetEntryByID :one
SELECT * FROM entries
WHERE id = ?;

-- name: GetEntriesByTag :many
SELECT * FROM entries
WHERE LOWER(TRIM(tag)) = LOWER(TRIM(?));

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
SELECT *
FROM entries
WHERE (
    review_count = 0
    OR datetime(last_reviewed_at, '+' || review_interval_days || ' days') <= datetime('now')
)
ORDER BY review_count ASC, last_reviewed_at ASC;

-- name: UpdateReview :exec
UPDATE entries
SET last_reviewed_at     = ?,
    review_interval_days = ?,
    ease_factor          = ?,
    review_count         = ?
WHERE id = ?;

-- name: CountEntriesByTag :many
SELECT tag, COUNT(*) AS count
FROM entries
GROUP BY tag
ORDER BY count DESC;

-- name: CountDueEntries :one
SELECT COUNT(*) FROM entries
WHERE (
    review_count = 0
    OR datetime(last_reviewed_at, '+' || review_interval_days || ' days') <= datetime('now')
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
    review_count         = 0;
