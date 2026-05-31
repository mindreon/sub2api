-- KVoucher PIN purchase orders (independent from payment_orders)

CREATE TABLE IF NOT EXISTS voucher_products (
    id BIGSERIAL PRIMARY KEY,
    kv_product_id BIGINT,
    name VARCHAR(128) NOT NULL,
    denomination DECIMAL(20,2) NOT NULL,
    wholesale_price DECIMAL(20,2) NOT NULL DEFAULT 0,
    retail_price DECIMAL(20,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'MYR',
    stock_available INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS voucher_products_kv_product_id_key
    ON voucher_products (kv_product_id) WHERE kv_product_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS voucher_products_denomination_idx ON voucher_products (denomination);
CREATE INDEX IF NOT EXISTS voucher_products_is_active_idx ON voucher_products (is_active);

CREATE TABLE IF NOT EXISTS voucher_orders (
    id BIGSERIAL PRIMARY KEY,
    order_no VARCHAR(64) NOT NULL UNIQUE,
    user_id BIGINT NOT NULL REFERENCES users(id),
    user_email VARCHAR(255) NOT NULL,
    user_name VARCHAR(100) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'pending_payment',
    product_id BIGINT NOT NULL,
    kv_product_id BIGINT,
    product_name VARCHAR(128) NOT NULL,
    denomination DECIMAL(20,2) NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 1,
    unit_price DECIMAL(20,2) NOT NULL,
    subtotal DECIMAL(20,2) NOT NULL,
    fee_amount DECIMAL(20,2) NOT NULL DEFAULT 0,
    total_amount DECIMAL(20,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'MYR',
    payment_ref VARCHAR(128),
    payment_proof_path VARCHAR(512),
    bank_account_id INTEGER,
    kv_retrieve_reference VARCHAR(128) NOT NULL,
    idempotency_key VARCHAR(128),
    reject_reason TEXT,
    fulfill_error TEXT,
    expires_at TIMESTAMPTZ NOT NULL,
    verified_at TIMESTAMPTZ,
    fulfilled_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    client_ip VARCHAR(50) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS voucher_orders_user_id_idx ON voucher_orders (user_id);
CREATE INDEX IF NOT EXISTS voucher_orders_status_idx ON voucher_orders (status);
CREATE INDEX IF NOT EXISTS voucher_orders_created_at_idx ON voucher_orders (created_at);
CREATE UNIQUE INDEX IF NOT EXISTS voucher_orders_idempotency_key_key
    ON voucher_orders (idempotency_key)
    WHERE idempotency_key IS NOT NULL AND idempotency_key <> '';

CREATE TABLE IF NOT EXISTS voucher_pin_deliveries (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL REFERENCES voucher_orders(id) ON DELETE CASCADE,
    pin_code_enc TEXT NOT NULL,
    serial VARCHAR(128) NOT NULL DEFAULT '',
    denomination DECIMAL(20,2) NOT NULL,
    expires_at TIMESTAMPTZ,
    delivered_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS voucher_pin_deliveries_order_id_idx ON voucher_pin_deliveries (order_id);

CREATE TABLE IF NOT EXISTS voucher_audit_logs (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL REFERENCES voucher_orders(id) ON DELETE CASCADE,
    action VARCHAR(64) NOT NULL,
    operator VARCHAR(128) NOT NULL,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS voucher_audit_logs_order_id_idx ON voucher_audit_logs (order_id);
CREATE INDEX IF NOT EXISTS voucher_audit_logs_created_at_idx ON voucher_audit_logs (created_at);
