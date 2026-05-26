ALTER TABLE commission_ledger
    ADD COLUMN IF NOT EXISTS settlement_reference_no VARCHAR(120) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS settlement_note TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS settled_by_user_id BIGINT NULL REFERENCES users(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_commission_ledger_settled_by_user_id
    ON commission_ledger(settled_by_user_id);
