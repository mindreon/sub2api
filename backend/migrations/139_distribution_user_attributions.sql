CREATE TABLE IF NOT EXISTS user_attributions (
    user_id BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    channel_org_id BIGINT NOT NULL REFERENCES channel_organizations(id) ON DELETE RESTRICT,
    referrer_member_id BIGINT NULL REFERENCES channel_members(id) ON DELETE SET NULL,
    promotion_link_id BIGINT NULL REFERENCES promotion_links(id) ON DELETE SET NULL,
    bound_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    bound_source VARCHAR(20) NOT NULL DEFAULT 'registration',
    bound_by VARCHAR(20) NOT NULL DEFAULT 'system',
    audit_id BIGINT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT user_attributions_bound_source_check CHECK (bound_source IN ('registration', 'oauth', 'manual')),
    CONSTRAINT user_attributions_bound_by_check CHECK (bound_by IN ('system', 'admin', 'import'))
);

CREATE INDEX IF NOT EXISTS idx_user_attributions_channel_org_id
    ON user_attributions(channel_org_id);

CREATE INDEX IF NOT EXISTS idx_user_attributions_referrer_member_id
    ON user_attributions(referrer_member_id);

CREATE INDEX IF NOT EXISTS idx_user_attributions_promotion_link_id
    ON user_attributions(promotion_link_id);

