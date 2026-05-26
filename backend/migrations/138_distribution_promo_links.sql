CREATE TABLE IF NOT EXISTS promotion_links (
    id BIGSERIAL PRIMARY KEY,
    channel_org_id BIGINT NOT NULL REFERENCES channel_organizations(id) ON DELETE CASCADE,
    member_id BIGINT NOT NULL REFERENCES channel_members(id) ON DELETE CASCADE,
    code VARCHAR(64) NOT NULL UNIQUE,
    target_type VARCHAR(20) NOT NULL DEFAULT 'registration',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT promotion_links_target_type_check CHECK (target_type IN ('registration', 'oauth', 'manual')),
    CONSTRAINT promotion_links_status_check CHECK (status IN ('active', 'inactive', 'disabled'))
);

CREATE INDEX IF NOT EXISTS idx_promotion_links_channel_org_id
    ON promotion_links(channel_org_id);

CREATE INDEX IF NOT EXISTS idx_promotion_links_member_id
    ON promotion_links(member_id);

