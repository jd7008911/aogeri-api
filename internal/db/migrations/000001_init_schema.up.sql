-- internal/db/migrations/000001_init_schema.up.sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    wallet_address VARCHAR(255) UNIQUE,
    two_factor_secret VARCHAR(255),
    two_factor_enabled BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    failed_login_attempts INTEGER DEFAULT 0,
    locked_until TIMESTAMP,
    last_login TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- User profiles
CREATE TABLE user_profiles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    username VARCHAR(100) UNIQUE,
    full_name VARCHAR(255),
    avatar_url TEXT,
    country VARCHAR(100),
    timezone VARCHAR(50),
    notifications_enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tokens (AOG, BNB, etc.)
CREATE TABLE tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    symbol VARCHAR(10) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    contract_address VARCHAR(255),
    decimals INTEGER DEFAULT 18,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Staking positions
CREATE TABLE stakes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    token_id UUID REFERENCES tokens(id),
    amount DECIMAL(36, 18) NOT NULL,
    apy DECIMAL(10, 4) DEFAULT 0,
    start_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    end_date TIMESTAMP,
    status VARCHAR(20) DEFAULT 'active', -- active, unstaked, claimed
    auto_compound BOOLEAN DEFAULT FALSE,
    rewards_claimed DECIMAL(36, 18) DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Liquidity pools
CREATE TABLE liquidity_pools (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    token0_id UUID REFERENCES tokens(id),
    token1_id UUID REFERENCES tokens(id),
    total_liquidity DECIMAL(36, 18) DEFAULT 0,
    apr DECIMAL(10, 4) DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- User liquidity positions
CREATE TABLE user_liquidity (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    pool_id UUID REFERENCES liquidity_pools(id),
    amount0 DECIMAL(36, 18) NOT NULL,
    amount1 DECIMAL(36, 18) NOT NULL,
    shares DECIMAL(36, 18) NOT NULL,
    apr_earned DECIMAL(10, 4) DEFAULT 0,
    status VARCHAR(20) DEFAULT 'active', -- active, withdrawn
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Governance proposals
CREATE TABLE governance_proposals (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    proposer_id UUID REFERENCES users(id),
    proposal_type VARCHAR(50) NOT NULL, -- fee_change, protocol_upgrade, etc.
    status VARCHAR(20) DEFAULT 'pending', -- pending, active, passed, rejected, executed
    voting_start TIMESTAMP,
    voting_end TIMESTAMP NOT NULL,
    quorum DECIMAL(5, 2) DEFAULT 50.0,
    threshold DECIMAL(5, 2) DEFAULT 50.0,
    for_votes DECIMAL(36, 18) DEFAULT 0,
    against_votes DECIMAL(36, 18) DEFAULT 0,
    abstain_votes DECIMAL(36, 18) DEFAULT 0,
    total_votes DECIMAL(36, 18) DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- User votes
CREATE TABLE user_votes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    proposal_id UUID REFERENCES governance_proposals(id) ON DELETE CASCADE,
    vote_power DECIMAL(36, 18) NOT NULL,
    vote_choice VARCHAR(10) NOT NULL, -- for, against, abstain
    voted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, proposal_id)
);

-- Assets tracking
CREATE TABLE assets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    token_id UUID REFERENCES tokens(id),
    total_value_locked DECIMAL(36, 18) DEFAULT 0,
    circulating_supply DECIMAL(36, 18) DEFAULT 0,
    market_price DECIMAL(36, 18) DEFAULT 0,
    price_change_24h DECIMAL(10, 4) DEFAULT 0,
    volume_24h DECIMAL(36, 18) DEFAULT 0,
    recorded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Security monitors
CREATE TABLE security_monitors (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    metric_name VARCHAR(100) NOT NULL,
    metric_value VARCHAR(255) NOT NULL,
    severity VARCHAR(20) DEFAULT 'info', -- info, warning, critical
    status VARCHAR(20) DEFAULT 'active', -- active, resolved
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- User sessions
CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    access_token_hash VARCHAR(255) NOT NULL,
    refresh_token_hash VARCHAR(255) NOT NULL,
    user_agent TEXT,
    ip_address INET,
    expires_at TIMESTAMP NOT NULL,
    revoked BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Audit logs
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id VARCHAR(100),
    details JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_wallet ON users(wallet_address);
CREATE INDEX idx_stakes_user ON stakes(user_id);
CREATE INDEX idx_stakes_status ON stakes(status);
CREATE INDEX idx_proposals_status ON governance_proposals(status);
CREATE INDEX idx_proposals_voting ON governance_proposals(voting_end);
CREATE INDEX idx_votes_proposal ON user_votes(proposal_id);
CREATE INDEX idx_sessions_user ON user_sessions(user_id);
CREATE INDEX idx_sessions_expires ON user_sessions(expires_at);
CREATE INDEX idx_audit_logs_user ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_created ON audit_logs(created_at);