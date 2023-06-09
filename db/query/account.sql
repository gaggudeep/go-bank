-- name: CreateAccount :one
INSERT INTO accounts(owner_name, balance, currency)
VALUES($1, $2, $3)
RETURNING *;

-- name: GetAccount :one
SELECT * FROM accounts
WHERE id = $1
FOR NO KEY UPDATE;

-- name: GetAccounts :many
SELECT * FROM accounts
WHERE owner_name = $1
ORDER BY id
LIMIT $2
OFFSET $3;

-- name: AddToAccountBalance :one
UPDATE accounts
SET balance = balance + sqlc.arg(amount)
WHERE id = $1
RETURNING *;

-- name: DeleteAccount :exec
DELETE FROM accounts
WHERE id = $1;