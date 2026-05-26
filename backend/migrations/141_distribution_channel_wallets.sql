CREATE TABLE IF NOT EXISTS channel_wallets (
    channel_org_id BIGINT PRIMARY KEY REFERENCES channel_organizations(id) ON DELETE CASCADE,
    prepaid_balance DECIMAL(20,8) NOT NULL DEFAULT 0,
    commission_reserved DECIMAL(20,8) NOT NULL DEFAULT 0,
    total_recharged DECIMAL(20,8) NOT NULL DEFAULT 0,
    total_consumed DECIMAL(20,8) NOT NULL DEFAULT 0,
    warning_threshold DECIMAL(20,8) NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT channel_wallets_status_check CHECK (status IN ('active', 'inactive', 'disabled'))
);

