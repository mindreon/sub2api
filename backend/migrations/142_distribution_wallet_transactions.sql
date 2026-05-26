CREATE TABLE IF NOT EXISTS channel_wallet_transactions (
    id BIGSERIAL PRIMARY KEY,
    channel_org_id BIGINT NOT NULL REFERENCES channel_organizations(id) ON DELETE CASCADE,
    transaction_type VARCHAR(32) NOT NULL,
    amount DECIMAL(20,8) NOT NULL,
    prepaid_balance_before DECIMAL(20,8) NOT NULL DEFAULT 0,
    prepaid_balance_after DECIMAL(20,8) NOT NULL DEFAULT 0,
    commission_reserved_before DECIMAL(20,8) NOT NULL DEFAULT 0,
    commission_reserved_after DECIMAL(20,8) NOT NULL DEFAULT 0,
    reference_no VARCHAR(120) NOT NULL DEFAULT '',
    note TEXT NOT NULL DEFAULT '',
    operator_user_id BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT channel_wallet_transactions_type_check CHECK (
        transaction_type IN (
            'recharge',
            'consume',
            'commission_reserve',
            'commission_release',
            'commission_settle',
            'commission_deduct',
            'commission_refund'
        )
    )
);

CREATE INDEX IF NOT EXISTS idx_channel_wallet_transactions_channel_org_id
    ON channel_wallet_transactions(channel_org_id);

CREATE INDEX IF NOT EXISTS idx_channel_wallet_transactions_type
    ON channel_wallet_transactions(transaction_type);

CREATE INDEX IF NOT EXISTS idx_channel_wallet_transactions_created_at
    ON channel_wallet_transactions(created_at DESC, id DESC);
