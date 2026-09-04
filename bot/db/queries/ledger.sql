-- name: CreateAccount :one
INSERT INTO accounts (player_id, type, created_at)
VALUES ($1, $2, NOW())
RETURNING id, player_id, type AS account_type, created_at;

-- name: GetAccountByType :one
-- Fetches a specific account type for a player (e.g., 'WALLET', 'BANK', 'ESCROW').
-- Fixed: account_type → type
SELECT id, player_id, type AS account_type, created_at
FROM accounts
WHERE player_id = @player_id AND type = @account_type;

-- name: GetAccountsByPlayer :many
-- Fixed: account_type → type
SELECT id, player_id, type AS account_type, created_at
FROM accounts
WHERE player_id = $1;

-- name: GetTreasuryAccount :one
-- Fixed: account_type → type
SELECT id, player_id, type AS account_type, created_at
FROM accounts
WHERE type = 'TREASURY'
LIMIT 1;

-- name: PostTransaction :one
-- Append-only transaction record. Amount is always positive; direction is determined by entry_type.
INSERT INTO transactions (account_id, amount, entry_type, transaction_type, reference_id, description, created_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW())
RETURNING id, account_id, amount, entry_type, transaction_type, reference_id, description, created_at;

-- name: GetAccountBalance :one
-- Computes the exact current balance. Explicit ::numeric casting ensures sqlc generates pgtype.Numeric.
SELECT COALESCE(SUM(CASE WHEN entry_type = 'CREDIT' THEN amount ELSE 0 END), 0)::numeric - 
       COALESCE(SUM(CASE WHEN entry_type = 'DEBIT' THEN amount ELSE 0 END), 0)::numeric AS balance
FROM transactions
WHERE account_id = $1;

-- name: GetAccountBalanceAndChange :one
-- Computes the current balance and the net change over the last 24 hours for a specific account.
-- Uses COALESCE to ensure zero is returned if no transactions exist, preventing NULL propagation.
SELECT 
    COALESCE(SUM(CASE WHEN entry_type = 'CREDIT' THEN amount ELSE -amount END), 0)::numeric AS current_balance,
    COALESCE(SUM(CASE 
        WHEN entry_type = 'CREDIT' AND created_at >= NOW() - INTERVAL '24 hours' THEN amount 
        WHEN entry_type = 'DEBIT' AND created_at >= NOW() - INTERVAL '24 hours' THEN -amount 
        ELSE 0 
    END), 0)::numeric AS change_24h
FROM transactions
WHERE account_id = @account_id;

-- name: GetTransactionHistory :many
SELECT id, account_id, amount, entry_type, transaction_type, reference_id, description, created_at
FROM transactions
WHERE account_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetPlayerWealthRank :one
-- Calculates the player's wealth percentile relative to all active players.
-- Note: For Layer 3+, this should be replaced by a query against the pre-computed `economic_snapshots` 
-- table to avoid full-table scans on the transactions table at scale.
WITH player_totals AS (
    SELECT 
        a.player_id,
        SUM(CASE WHEN t.entry_type = 'CREDIT' THEN t.amount ELSE -t.amount END) AS net_worth
    FROM accounts a
    JOIN transactions t ON t.account_id = a.id
    GROUP BY a.player_id
),
ranked AS (
    SELECT 
        player_id,
        PERCENT_RANK() OVER (ORDER BY net_worth) * 100 AS rank_percent
    FROM player_totals
)
SELECT COALESCE(FLOOR(rank_percent)::integer, 0) AS rank_percent
FROM ranked
WHERE player_id = @player_id;

-- name: GetLatestEconomicSnapshot :one
-- Fetches the most recent economic snapshot by economic_day.
SELECT id, economic_day, metrics, created_at
FROM economic_snapshots
ORDER BY economic_day DESC
LIMIT 1;

-- name: GetPreviousEconomicSnapshot :one
-- Fetches the snapshot immediately before the given economic day.
SELECT id, economic_day, metrics, created_at
FROM economic_snapshots
WHERE economic_day < @economic_day
ORDER BY economic_day DESC
LIMIT 1;

-- name: GetUnpaidWages :many
-- Identifies labor records that do not have a corresponding WAGE_PAYMENT transaction.
-- This is used by the wage reconciliation job to detect settlement failures.
SELECT dl.player_id AS employee_id, dl.business_id, dl.hours_worked, 
       (e.wage_rate * dl.hours_worked) AS amount
FROM daily_labor dl
JOIN employment e ON dl.business_id = e.business_id AND dl.player_id = e.employee_id
LEFT JOIN transactions t ON t.reference_id = dl.player_id 
    AND t.transaction_type = 'WAGE_PAYMENT'
    AND t.created_at >= dl.created_at
WHERE t.id IS NULL;

-- name: GetHourlyTaxMetrics :one
-- Aggregates tax collection metrics from the past hour.
SELECT 
    COALESCE(SUM(t.amount), 0)::numeric AS total_tax_collected,
    COUNT(*)::bigint AS transaction_count
FROM transactions t
WHERE t.transaction_type IN ('TAX_COLLECTION', 'FEE')
  AND t.created_at >= NOW() - INTERVAL '1 hour';