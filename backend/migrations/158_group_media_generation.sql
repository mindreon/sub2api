-- Migration: 158_group_media_generation
-- 分组级多模态（视频/音频/图片）异步生成开关与独立倍率（仿 allow_image_generation）。

ALTER TABLE groups
    ADD COLUMN IF NOT EXISTS allow_media_generation BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS media_rate_independent BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS media_rate_multiplier DECIMAL(10,4) NOT NULL DEFAULT 1.0;

COMMENT ON COLUMN groups.allow_media_generation IS '是否允许该分组 API Key 调用 /v1/videos 等多模态异步接口';
COMMENT ON COLUMN groups.media_rate_independent IS '多模态是否使用独立倍率；false 表示共享分组有效倍率';
COMMENT ON COLUMN groups.media_rate_multiplier IS '多模态独立倍率，仅 media_rate_independent=true 时生效';
