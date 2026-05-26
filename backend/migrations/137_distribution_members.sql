CREATE TABLE IF NOT EXISTS channel_members (
    id BIGSERIAL PRIMARY KEY,
    channel_org_id BIGINT NOT NULL REFERENCES channel_organizations(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_type VARCHAR(20) NOT NULL,
    parent_member_id BIGINT NULL REFERENCES channel_members(id) ON DELETE SET NULL,
    level_code VARCHAR(20) NOT NULL DEFAULT '',
    commission_rate DECIMAL(10,4) NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT channel_members_role_type_check CHECK (role_type IN ('manager', 'agent', 'kol1', 'kol2')),
    CONSTRAINT channel_members_status_check CHECK (status IN ('active', 'inactive', 'disabled')),
    CONSTRAINT channel_members_user_role_unique UNIQUE (user_id, channel_org_id, role_type)
);

CREATE INDEX IF NOT EXISTS idx_channel_members_channel_org_id
    ON channel_members(channel_org_id);

CREATE INDEX IF NOT EXISTS idx_channel_members_parent_member_id
    ON channel_members(parent_member_id);

CREATE INDEX IF NOT EXISTS idx_channel_members_user_id
    ON channel_members(user_id);

