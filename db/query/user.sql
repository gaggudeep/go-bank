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

