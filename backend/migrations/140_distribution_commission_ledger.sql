CREATE TABLE IF NOT EXISTS commission_ledger (
    id BIGSERIAL PRIMARY KEY,
    channel_org_id BIGINT NOT NULL REFERENCES channel_organizations(id) ON DELETE CASCADE,
    member_id BIGINT NOT NULL REFERENCES channel_members(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    usage_log_id BIGINT NULL REFERENCES usage_logs(id) ON DELETE SET NULL,
    commission_type VARCHAR(30) NOT NULL,
    base_amount DECIMAL(20,8) NOT NULL,
    rate DECIMAL(10,4) NOT NULL DEFAULT 0,
    amount DECIMAL(20,8) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    settlement_method VARCHAR(20) NOT NULL DEFAULT 'balance',
    frozen_until TIMESTAMPTZ NULL,
    settled_at TIMESTAMPTZ NULL,
    reversed_from_id BIGINT NULL REFERENCES commission_ledger(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT commission_ledger_status_check CHECK (status IN ('pending', 'frozen', 'available', 'settled', 'reversed', 'cancelled')),
    CONSTRAINT commission_ledger_settlement_method_check CHECK (settlement_method IN ('balance', 'auto', 'manual', 'offline'))
);

CREATE INDEX IF NOT EXISTS idx_commission_ledger_channel_org_id
    ON commission_ledger(channel_org_id);

CREATE INDEX IF NOT EXISTS idx_commission_ledger_member_id
    ON commission_ledger(member_id);

CREATE INDEX IF NOT EXISTS idx_commission_ledger_user_id
    ON commission_ledger(user_id);

CREATE INDEX IF NOT EXISTS idx_commission_ledger_usage_log_id
    ON commission_ledger(usage_log_id);

CREATE INDEX IF NOT EXISTS idx_commission_ledger_status
    ON commission_ledger(status);

CREATE INDEX IF NOT EXISTS idx_commission_ledger_frozen_until
    ON commission_ledger(frozen_until);

