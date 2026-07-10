-- Migration: 136_media_billing
-- 多模态异步计费：任务表 + 额度预扣台账

CREATE TABLE IF NOT EXISTS media_generation_tasks (
    id                BIGSERIAL PRIMARY KEY,
    task_id           VARCHAR(64) NOT NULL UNIQUE,
    upstream_task_id  VARCHAR(128),
    user_id           BIGINT NOT NULL,
    api_key_id        BIGINT NOT NULL,
    account_id        BIGINT,
    group_id          BIGINT,
    model             VARCHAR(100) NOT NULL,
    media_type        VARCHAR(16) NOT NULL,
    status            VARCHAR(20) NOT NULL DEFAULT 'pending',
    billing_metric    VARCHAR(32),
    reserved_cost     DECIMAL(20,10) NOT NULL DEFAULT 0,
    actual_cost       DECIMAL(20,10),
    rate_multiplier   DECIMAL(10,4) NOT NULL DEFAULT 1,
    billing_currency  VARCHAR(8) NOT NULL DEFAULT 'USD',
    request_params    JSONB,
    upstream_usage    JSONB,
    poll_attempts     INT NOT NULL DEFAULT 0,
    expires_at        TIMESTAMPTZ NOT NULL,
    settled_at        TIMESTAMPTZ,
    error_message     TEXT,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_media_generation_tasks_user_status
    ON media_generation_tasks (user_id, status);
CREATE INDEX IF NOT EXISTS idx_media_generation_tasks_status_expires
    ON media_generation_tasks (status, expires_at);
CREATE INDEX IF NOT EXISTS idx_media_generation_tasks_created_at
    ON media_generation_tasks (created_at);

COMMENT ON TABLE media_generation_tasks IS '多模态异步生成任务（视频/音频/图片）';
COMMENT ON COLUMN media_generation_tasks.task_id IS '对外任务 ID';
COMMENT ON COLUMN media_generation_tasks.reserved_cost IS '预扣金额（计费币种，预估上限）';

CREATE TABLE IF NOT EXISTS media_quota_holds (
    id         BIGSERIAL PRIMARY KEY,
    hold_id    VARCHAR(64) NOT NULL UNIQUE,
    task_id    VARCHAR(64) NOT NULL UNIQUE,
    user_id    BIGINT NOT NULL,
    amount     DECIMAL(20,10) NOT NULL,
    currency   VARCHAR(8) NOT NULL DEFAULT 'USD',
    status     VARCHAR(16) NOT NULL DEFAULT 'held',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_media_quota_holds_user_status
    ON media_quota_holds (user_id, status);

COMMENT ON TABLE media_quota_holds IS '多模态任务额度预扣台账';
COMMENT ON COLUMN media_quota_holds.status IS 'held=冻结中, settled=已结算, released=已释放';
