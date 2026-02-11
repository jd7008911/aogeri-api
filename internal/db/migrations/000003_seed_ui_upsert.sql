-- internal/db/migrations/000003_seed_ui_upsert.sql

-- Upsert assets for AOG
WITH t AS (SELECT id AS tid FROM tokens WHERE symbol='AOG')
UPDATE assets SET
    total_value_locked = 123500000,
    market_price = 10.00,
    price_change_24h = 56.4,
    volume_24h = 5000000,
    recorded_at = CURRENT_TIMESTAMP
FROM t
WHERE assets.token_id = t.tid;

INSERT INTO assets (token_id, total_value_locked, market_price, price_change_24h, volume_24h, recorded_at)
SELECT t.tid, 123500000, 10.00, 56.4, 5000000, CURRENT_TIMESTAMP
FROM (SELECT id AS tid FROM tokens WHERE symbol='AOG') t
WHERE NOT EXISTS (SELECT 1 FROM assets WHERE token_id = t.tid);

-- Insert sample stakes if not existed (by amount and start_date)
DO $$
DECLARE uid uuid;
DECLARE tid uuid;
BEGIN
    SELECT id INTO uid FROM users WHERE email='dev@example.com' LIMIT 1;
    SELECT id INTO tid FROM tokens WHERE symbol='AOG' LIMIT 1;
    IF uid IS NULL OR tid IS NULL THEN
        RAISE NOTICE 'seed skipped: missing user or token';
        RETURN;
    END IF;

    -- Insert only if no active stake exists with same amount
    IF NOT EXISTS (SELECT 1 FROM stakes WHERE user_id=uid AND token_id=tid AND amount=10.5 AND status='active') THEN
        INSERT INTO stakes (user_id, token_id, amount, apy, start_date, end_date, status, auto_compound, rewards_claimed)
        VALUES (uid, tid, 10.5, 33.29, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP + INTERVAL '30 days', 'active', false, 0);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM stakes WHERE user_id=uid AND token_id=tid AND amount=250 AND status='active') THEN
        INSERT INTO stakes (user_id, token_id, amount, apy, start_date, end_date, status, auto_compound, rewards_claimed)
        VALUES (uid, tid, 250, 33.29, CURRENT_TIMESTAMP - INTERVAL '5 days', CURRENT_TIMESTAMP + INTERVAL '25 days', 'active', true, 0);
    END IF;
END $$;

-- Liquidity pool upsert by name
INSERT INTO liquidity_pools (name, token0_id, token1_id, total_liquidity, apr, is_active, created_at)
SELECT 'AOG-BNB', (SELECT id FROM tokens WHERE symbol='AOG'), (SELECT id FROM tokens WHERE symbol='BNB'), 500000, 12.0, true, CURRENT_TIMESTAMP
WHERE NOT EXISTS (SELECT 1 FROM liquidity_pools WHERE name='AOG-BNB');

-- Security monitors insert if missing
INSERT INTO security_monitors (metric_name, metric_value, severity, status, created_at)
SELECT 'suspicious-withdrawals','25 alerts in last 24h','warning','active',CURRENT_TIMESTAMP
WHERE NOT EXISTS (SELECT 1 FROM security_monitors WHERE metric_name='suspicious-withdrawals');

INSERT INTO security_monitors (metric_name, metric_value, severity, status, created_at)
SELECT 'multi-sig-health','95% approvals','info','active',CURRENT_TIMESTAMP
WHERE NOT EXISTS (SELECT 1 FROM security_monitors WHERE metric_name='multi-sig-health');

-- Governance proposals insert if missing
INSERT INTO governance_proposals (title, description, proposer_id, proposal_type, status, voting_start, voting_end, created_at)
SELECT 'Increase AOG staking APY', 'Proposal to increase APY for AOG staking to attract liquidity', (SELECT id FROM users WHERE email='dev@example.com'), 'fee_change', 'active', CURRENT_TIMESTAMP - INTERVAL '1 day', CURRENT_TIMESTAMP + INTERVAL '6 days', CURRENT_TIMESTAMP
WHERE NOT EXISTS (SELECT 1 FROM governance_proposals WHERE title='Increase AOG staking APY');

INSERT INTO governance_proposals (title, description, proposer_id, proposal_type, status, voting_start, voting_end, created_at)
SELECT 'Add new rewards pool', 'Create a new rewards pool for liquidity providers', (SELECT id FROM users WHERE email='dev@example.com'), 'protocol_upgrade', 'pending', NULL, CURRENT_TIMESTAMP + INTERVAL '30 days', CURRENT_TIMESTAMP
WHERE NOT EXISTS (SELECT 1 FROM governance_proposals WHERE title='Add new rewards pool');

-- User profile upsert: create if not exists
INSERT INTO user_profiles (user_id, username, full_name, notifications_enabled, created_at)
SELECT id, 'devuser', 'Dev Example', true, CURRENT_TIMESTAMP FROM users WHERE email='dev@example.com'
AND NOT EXISTS (SELECT 1 FROM user_profiles WHERE user_id = (SELECT id FROM users WHERE email='dev@example.com'));
