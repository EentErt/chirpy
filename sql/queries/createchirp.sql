-- name: CreateChirp :one
INSERT INTO chirps(id, created_at, updated_at, body, user_id)
VALUES (
    gen_random_uuid(),
    NOW(),
    nOW(),
    $1,
    $2
)
RETURNING *;