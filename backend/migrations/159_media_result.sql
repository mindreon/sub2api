-- Migration: 159_media_result
-- 多模态异步任务结果回传：记录生成视频链接与自有存储对象键。

ALTER TABLE media_generation_tasks
    ADD COLUMN IF NOT EXISTS result_url TEXT,
    ADD COLUMN IF NOT EXISTS result_storage_key VARCHAR(255);

COMMENT ON COLUMN media_generation_tasks.result_url IS '对客户端展示的结果视频链接（转存后为自有存储链接，否则为上游直链）';
COMMENT ON COLUMN media_generation_tasks.result_storage_key IS '自有对象存储的对象键（转存成功时写入，用于按需重新签发链接）';
