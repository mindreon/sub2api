CREATE TABLE IF NOT EXISTS catalog_models (
    id              BIGSERIAL PRIMARY KEY,
    model_id        VARCHAR(200) NOT NULL,
    name            VARCHAR(200) NOT NULL,
    vendor          VARCHAR(100) NOT NULL DEFAULT '',
    category        VARCHAR(50)  NOT NULL DEFAULT 'chat',
    description     TEXT         NOT NULL DEFAULT '',
    tags            JSONB        NOT NULL DEFAULT '[]'::jsonb,
    doc_url         TEXT         NOT NULL DEFAULT '',
    icon_url        TEXT         NOT NULL DEFAULT '',
    context_window  BIGINT       NOT NULL DEFAULT 0,
    max_output_tokens BIGINT     NOT NULL DEFAULT 0,
    input_modalities  JSONB      NOT NULL DEFAULT '[]'::jsonb,
    output_modalities JSONB      NOT NULL DEFAULT '[]'::jsonb,
    features          JSONB      NOT NULL DEFAULT '[]'::jsonb,
    input_price       DOUBLE PRECISION NOT NULL DEFAULT 0,
    output_price      DOUBLE PRECISION NOT NULL DEFAULT 0,
    cache_write_price DOUBLE PRECISION,
    cache_read_price  DOUBLE PRECISION,
    currency        VARCHAR(10)  NOT NULL DEFAULT 'USD',
    is_enabled      BOOLEAN      NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS catalog_models_model_id_key ON catalog_models(model_id);
CREATE INDEX IF NOT EXISTS idx_catalog_models_is_enabled ON catalog_models(is_enabled);
CREATE INDEX IF NOT EXISTS idx_catalog_models_vendor ON catalog_models(vendor);
CREATE INDEX IF NOT EXISTS idx_catalog_models_category ON catalog_models(category);
