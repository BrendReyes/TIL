-- name: CreateEntry :one
INSERT INTO entries (body, tag) 
VALUES (
    ?,
    ?
)
RETURNING *;