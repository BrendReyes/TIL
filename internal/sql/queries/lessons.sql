-- name: CreateLesson :one
INSERT INTO lessons (body, tag) 
VALUES (
    ?,
    ?
)
RETURNING *;