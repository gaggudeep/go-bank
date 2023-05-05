-- name: CreateTransaction :one
INSERT INTO transactions(account_id, amount)
VALUES($1, $2)
RETURNING *;

-- name: GetTransaction :one
SELECT * FROM transactions
WHERE id = $1;

-- name: DeleteTransaction :exec
DELETE FROM transactions
WHERE id = $1;