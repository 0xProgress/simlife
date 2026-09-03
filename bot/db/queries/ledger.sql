-- name: CreateAccount :one
INSERT INTO accounts (player_id, type)
VALUES ($1, $2)
RETURNING *;

-- name: GetAccountsByPlayer :many
SELECT * FROM accounts
WHERE player_id = $1;

-- name: GetTreasuryAccount :one
SELECT * FROM accounts
WHERE type = 'TREASURY'
LIMIT 1;

-- name: PostTransaction :one
INSERT INTO transactions (account_id, amount, entry_type, transaction_type, reference_id, description)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetAccountBalance :one
SELECT COALESCE(SUM(CASE WHEN entry_type = 'CREDIT' THEN amount ELSE 0 END), 0) - 
       COALESCE(SUM(CASE WHEN entry_type = 'DEBIT' THEN amount ELSE 0 END), 0) AS balance
FROM transactions
WHERE account_id = $1;

-- name: GetTransactionHistory :many
SELECT * FROM transactions
WHERE account_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;