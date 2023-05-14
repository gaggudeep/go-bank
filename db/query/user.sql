-- name: CreateUser :one
INSERT INTO users (
   username,
   hashed_password,
   name,
   email
) VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetUser :one
SELECT * FROM users
where username = $1;

-- name: UpdateUser :one
UPDATE users
SET
    hashed_password = COALESCE(sqlc.narg(hashed_password), hashed_password),
    name = COALESCE(sqlc.narg(name), name),
    email = COALESCE(sqlc.narg(email), email),
    password_changed_at = CASE
        WHEN sqlc.narg(hashed_password) IS NOT NULL THEN NOW()
        ELSE password_changed_at
    END
WHERE
    username = sqlc.arg(username)
RETURNING *;