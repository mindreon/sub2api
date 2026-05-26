CREATE TABLE IF NOT EXISTS distribution_attribution_audits (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    previous_channel_org_id BIGINT NULL REFERENCES channel_organizations(id) ON DELETE SET NULL,
    previous_referrer_member_id BIGINT NULL REFERENCES channel_members(id) ON DELETE SET NULL,
    previous_promotion_link_id BIGINT NULL REFERENCES promotion_links(id) ON DELETE SET NULL,
    previous_bound_source VARCHAR(20) NULL,
    previous_bound_by VARCHAR(20) NULL,
    new_channel_org_id BIGINT NOT NULL REFERENCES channel_organizations(id) ON DELETE RESTRICT,
    new_referrer_member_id BIGINT NULL REFERENCES channel_members(id) ON DELETE SET NULL,
    new_promotion_link_id BIGINT NULL REFERENCES promotion_links(id) ON DELETE SET NULL,
    new_bound_source VARCHAR(20) NOT NULL DEFAULT 'manual',
    new_bound_by VARCHAR(20) NOT NULL DEFAULT 'admin',
    note VARCHAR(2000) NOT NULL DEFAULT '',
    operator_user_id BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT distribution_attribution_audits_previous_bound_source_check CHECK (
        previous_bound_source IS NULL OR previous_bound_source IN ('registration', 'oauth', 'manual')
    ),
    CONSTRAINT distribution_attribution_audits_previous_bound_by_check CHECK (
        previous_bound_by IS NULL OR previous_bound_by IN ('system', 'admin', 'import')
    ),
    CONSTRAINT distribution_attribution_audits_new_bound_source_check CHECK (
        new_bound_source IN ('registration', 'oauth', 'manual')
    ),
    CONSTRAINT distribution_attribution_audits_new_bound_by_check CHECK (
        new_bound_by IN ('system', 'admin', 'import')
    )
);

CREATE INDEX IF NOT EXISTS idx_distribution_attribution_audits_user_id
    ON distribution_attribution_audits(user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_distribution_attribution_audits_operator_user_id
    ON distribution_attribution_audits(operator_user_id, created_at DESC);
