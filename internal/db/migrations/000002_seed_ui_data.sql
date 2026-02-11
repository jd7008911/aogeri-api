-- internal/db/migrations/000002_seed_ui_data.sql

-- Seed Assets (match dashboard TVL and prices)
INSERT INTO assets (token_id, total_value_locked, market_price, price_change_24h, volume_24h, recorded_at)
VALUES (
    (SELECT id FROM tokens WHERE symbol='AOG'),
    123500000, -- $123.5M
    10.00,     -- market price per token
    56.4,      -- 24h price change
    5000000,   -- 24h volume
    CURRENT_TIMESTAMP
)
ON CONFLICT (token_id) DO UPDATE SET
    total_value_locked = EXCLUDED.total_value_locked,
    market_price = EXCLUDED.market_price,
    price_change_24h = EXCLUDED.price_change_24h,
    volume_24h = EXCLUDED.volume_24h,
    recorded_at = EXCLUDED.recorded_at;

-- Insert a few representative stakes for the sample user
DO $$
DECLARE
    uid uuid;
    tid uuid;
BEGIN
    SELECT id INTO uid FROM users WHERE email='dev@example.com' LIMIT 1;
    SELECT id INTO tid FROM tokens WHERE symbol='AOG' LIMIT 1;
    IF uid IS NULL OR tid IS NULL THEN
        RAISE NOTICE 'seed skipped: missing user or token';
        RETURN;
    END IF;

    INSERT INTO stakes (user_id, token_id, amount, apy, start_date, end_date, status, auto_compound, rewards_claimed)
    VALUES
    (uid, tid, 10.5, 33.29, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP + INTERVAL '30 days', 'active', false, 0),
    (uid, tid, 250.0, 33.29, CURRENT_TIMESTAMP - INTERVAL '5 days', CURRENT_TIMESTAMP + INTERVAL '25 days', 'active', true, 0),
    (uid, tid, 1000.0, 12.50, CURRENT_TIMESTAMP - INTERVAL '40 days', CURRENT_TIMESTAMP - INTERVAL '10 days', 'unstaked', false, 0);
END $$;

-- Liquidity pool example
INSERT INTO liquidity_pools (name, token0_id, token1_id, total_liquidity, apr, is_active, created_at)
VALUES (
    'AOG-BNB', (SELECT id FROM tokens WHERE symbol='AOG'), (SELECT id FROM tokens WHERE symbol='BNB'), 500000, 12.0, true, CURRENT_TIMESTAMP
)
ON CONFLICT (name) DO NOTHING;

-- Security monitors
INSERT INTO security_monitors (metric_name, metric_value, severity, status, created_at)
VALUES
('suspicious-withdrawals', '25 alerts in last 24h', 'warning', 'active', CURRENT_TIMESTAMP),
('multi-sig-health', '95% approvals', 'info', 'active', CURRENT_TIMESTAMP)
ON CONFLICT (metric_name) DO NOTHING;

-- Governance proposals
INSERT INTO governance_proposals (title, description, proposer_id, proposal_type, status, voting_start, voting_end)
VALUES
('Increase AOG staking APY', 'Proposal to increase APY for AOG staking to attract liquidity', (SELECT id FROM users WHERE email='dev@example.com'), 'fee_change', 'active', CURRENT_TIMESTAMP - INTERVAL '1 day', CURRENT_TIMESTAMP + INTERVAL '6 days'),
('Add new rewards pool', 'Create a new rewards pool for liquidity providers', (SELECT id FROM users WHERE email='dev@example.com'), 'protocol_upgrade', 'pending', NULL, CURRENT_TIMESTAMP + INTERVAL '30 days')
ON CONFLICT (title) DO NOTHING;

-- Example user profile
INSERT INTO user_profiles (user_id, username, full_name, notifications_enabled, created_at)
VALUES ((SELECT id FROM users WHERE email='dev@example.com'), 'devuser', 'Dev Example', true, CURRENT_TIMESTAMP)
ON CONFLICT (user_id) DO NOTHING;
