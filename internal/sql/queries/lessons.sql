-- name: CreateEntry :one
INSERT INTO entries (body, tag, created_at) 
VALUES (
    ?,
    ?,
    ?
)
RETURNING *;