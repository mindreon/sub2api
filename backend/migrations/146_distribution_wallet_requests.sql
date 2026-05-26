CREATE TABLE IF NOT EXISTS channel_wallet_requests (
    id BIGSERIAL PRIMARY KEY,
    channel_org_id BIGINT NOT NULL REFERENCES channel_organizations(id) ON DELETE CASCADE,
    request_type VARCHAR(32) NOT NULL,
    amount DECIMAL(20,8) NOT NULL,
    reference_no VARCHAR(120) NOT NULL DEFAULT '',
    note TEXT NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_by_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    reviewed_by_user_id BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    review_note TEXT NOT NULL DEFAULT '',
    reviewed_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT channel_wallet_requests_type_check CHECK (
        request_type IN ('recharge', 'refund')
    ),
    CONSTRAINT channel_wallet_requests_status_check CHECK (
        status IN ('pending', 'approved', 'rejected')
    )
);

CREATE INDEX IF NOT EXISTS idx_channel_wallet_requests_channel_org_id
    ON channel_wallet_requests(channel_org_id);

CREATE INDEX IF NOT EXISTS idx_channel_wallet_requests_type_status
    ON channel_wallet_requests(request_type, status);

CREATE INDEX IF NOT EXISTS idx_channel_wallet_requests_created_by_user_id
    ON channel_wallet_requests(created_by_user_id, created_at DESC);
