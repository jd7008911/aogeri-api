-- internal/db/migrations/000001_init_schema.down.sql
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS user_sessions;
DROP TABLE IF EXISTS security_monitors;
DROP TABLE IF EXISTS assets;
DROP TABLE IF EXISTS user_votes;
DROP TABLE IF EXISTS governance_proposals;
DROP TABLE IF EXISTS user_liquidity;
DROP TABLE IF EXISTS liquidity_pools;
DROP TABLE IF EXISTS stakes;
DROP TABLE IF EXISTS tokens;
DROP TABLE IF EXISTS user_profiles;
DROP TABLE IF EXISTS users;