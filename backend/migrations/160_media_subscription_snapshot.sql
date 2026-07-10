-- Store the subscription selected when an async media task is submitted.
-- Settlement may happen much later, so billing must not depend on live group
-- subscription state at poll time.
ALTER TABLE media_generation_tasks
    ADD COLUMN IF NOT EXISTS subscription_id BIGINT;

CREATE INDEX IF NOT EXISTS idx_media_generation_tasks_subscription_id
    ON media_generation_tasks (subscription_id);

COMMENT ON COLUMN media_generation_tasks.subscription_id IS
    '提交时快照的用户订阅 ID，用于异步结算保持订阅计费语义';
