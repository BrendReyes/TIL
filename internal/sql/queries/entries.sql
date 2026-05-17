-- name: CreateEntry :one
INSERT INTO entries (body, tag, created_at) 
VALUES (
    ?,
    ?,
    ?
)
RETURNING *;

-- name: ListAllEntry :many
SELECT * FROM entries;

-- name: DeleteEntry :exec
DELETE FROM entries
WHERE id = ?;