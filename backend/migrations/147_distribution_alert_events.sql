CREATE TABLE IF NOT EXISTS channel_alert_events (
    id BIGSERIAL PRIMARY KEY,
    channel_org_id BIGINT NOT NULL REFERENCES channel_organizations(id) ON DELETE CASCADE,
    alert_type VARCHAR(64) NOT NULL,
    severity VARCHAR(20) NOT NULL DEFAULT 'warning',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    details JSONB NOT NULL DEFAULT '{}'::jsonb,
    triggered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ NULL,
    last_observed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT channel_alert_events_type_check CHECK (
        alert_type IN (
            'low_balance',
            'balance_exhausted',
            'consumption_warning',
            'consumption_exhausted'
        )
    ),
    CONSTRAINT channel_alert_events_severity_check CHECK (
        severity IN ('info', 'warning', 'critical')
    ),
    CONSTRAINT channel_alert_events_status_check CHECK (
        status IN ('active', 'resolved')
    )
);

CREATE INDEX IF NOT EXISTS idx_channel_alert_events_channel_org_status
    ON channel_alert_events(channel_org_id, status, triggered_at DESC);

CREATE INDEX IF NOT EXISTS idx_channel_alert_events_type_status
    ON channel_alert_events(alert_type, status, triggered_at DESC);

CREATE UNIQUE INDEX IF NOT EXISTS idx_channel_alert_events_active_unique
    ON channel_alert_events(channel_org_id, alert_type)
    WHERE status = 'active';
