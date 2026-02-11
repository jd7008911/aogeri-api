-- internal/db/queries/stakes.sql
-- name: CreateStake :one
INSERT INTO stakes (user_id, token_id, amount, apy, end_date, auto_compound)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetUserStakes :many
SELECT s.*, t.symbol, t.name
FROM stakes s
JOIN tokens t ON s.token_id = t.id
WHERE s.user_id = $1 AND s.status = 'active'
ORDER BY s.created_at DESC;

-- name: GetStakeByID :one
SELECT s.*, t.symbol, t.name
FROM stakes s
JOIN tokens t ON s.token_id = t.id
WHERE s.id = $1;

-- name: UpdateStakeRewards :exec
UPDATE stakes 
SET rewards_claimed = rewards_claimed + $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: Unstake :exec
UPDATE stakes 
SET status = 'unstaked', end_date = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND user_id = $2;

-- name: GetTotalStakedValue :one
SELECT COALESCE(SUM(s.amount * a.market_price), 0)::decimal as total_value
FROM stakes s
JOIN assets a ON s.token_id = a.token_id
WHERE s.status = 'active';