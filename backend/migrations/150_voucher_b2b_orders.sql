-- KVoucher B2B replenishment orders (platform -> KVoucher wholesale)

CREATE TABLE IF NOT EXISTS voucher_b2b_orders (
    id BIGSERIAL PRIMARY KEY,
    kv_order_id BIGINT NOT NULL UNIQUE,
    order_no VARCHAR(64) NOT NULL UNIQUE,
    status VARCHAR(32) NOT NULL DEFAULT 'pending_payment',
    subtotal DECIMAL(20,2) NOT NULL,
    fee_amount DECIMAL(20,2) NOT NULL DEFAULT 0,
    total_amount DECIMAL(20,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'MYR',
    items_json JSONB NOT NULL DEFAULT '[]',
    payment_info_json JSONB,
    payment_ref VARCHAR(128),
    payment_proof_path VARCHAR(512),
    bank_account_id INTEGER,
    merchant_notes VARCHAR(500),
    idempotency_key VARCHAR(128),
    reject_reason TEXT,
    kv_last_request_id VARCHAR(64),
    created_by VARCHAR(128) NOT NULL DEFAULT 'admin',
    kv_last_synced_at TIMESTAMPTZ,
    verified_at TIMESTAMPTZ,
    pins_loaded_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS voucher_b2b_orders_status_idx ON voucher_b2b_orders (status);
CREATE INDEX IF NOT EXISTS voucher_b2b_orders_created_at_idx ON voucher_b2b_orders (created_at);
CREATE UNIQUE INDEX IF NOT EXISTS voucher_b2b_orders_idempotency_key_key
    ON voucher_b2b_orders (idempotency_key)
    WHERE idempotency_key IS NOT NULL AND idempotency_key <> '';

ALTER TABLE voucher_audit_logs
    ALTER COLUMN order_id DROP NOT NULL;

ALTER TABLE voucher_audit_logs
    DROP CONSTRAINT IF EXISTS voucher_audit_logs_order_id_fkey;

ALTER TABLE voucher_audit_logs
    ADD COLUMN IF NOT EXISTS b2b_order_id BIGINT REFERENCES voucher_b2b_orders(id) ON DELETE CASCADE;

ALTER TABLE voucher_audit_logs
    ADD CONSTRAINT voucher_audit_logs_order_id_fkey
    FOREIGN KEY (order_id) REFERENCES voucher_orders(id) ON DELETE CASCADE;

ALTER TABLE voucher_audit_logs
    DROP CONSTRAINT IF EXISTS voucher_audit_logs_subject_check;

ALTER TABLE voucher_audit_logs
    ADD CONSTRAINT voucher_audit_logs_subject_check CHECK (
        (order_id IS NOT NULL AND b2b_order_id IS NULL)
        OR (order_id IS NULL AND b2b_order_id IS NOT NULL)
    );

CREATE INDEX IF NOT EXISTS voucher_audit_logs_b2b_order_id_idx ON voucher_audit_logs (b2b_order_id);
