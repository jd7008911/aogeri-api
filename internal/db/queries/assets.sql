-- internal/db/queries/assets.sql
-- name: UpdateAssetPrice :exec
INSERT INTO assets (token_id, market_price, price_change_24h, volume_24h, recorded_at)
VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP)
ON CONFLICT (token_id) 
DO UPDATE SET 
    market_price = EXCLUDED.market_price,
    price_change_24h = EXCLUDED.price_change_24h,
    volume_24h = EXCLUDED.volume_24h,
    recorded_at = EXCLUDED.recorded_at;

-- name: GetAssetMetrics :one
SELECT 
    COALESCE(SUM(total_value_locked), 0) as total_tvl,
    COUNT(DISTINCT token_id) as active_tokens,
    COUNT(*) as total_assets
FROM assets;

-- name: GetTokenList :many
SELECT t.*, a.market_price, a.price_change_24h
FROM tokens t
LEFT JOIN assets a ON t.id = a.token_id
WHERE t.is_active = true
ORDER BY t.symbol;