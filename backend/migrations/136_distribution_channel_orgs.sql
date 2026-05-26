CREATE TABLE IF NOT EXISTS channel_organizations (
    id BIGSERIAL PRIMARY KEY,
    type VARCHAR(20) NOT NULL,
    name VARCHAR(255) NOT NULL,
    owner_user_id BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    config JSONB NOT NULL DEFAULT '{}'::jsonb,
    brand_config JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT channel_organizations_type_check CHECK (type IN ('platform', 'reseller', 'oem')),
    CONSTRAINT channel_organizations_status_check CHECK (status IN ('active', 'inactive', 'disabled'))
);

CREATE INDEX IF NOT EXISTS idx_channel_organizations_owner_user_id
    ON channel_organizations(owner_user_id);

CREATE INDEX IF NOT EXISTS idx_channel_organizations_type_status
    ON channel_organizations(type, status);

