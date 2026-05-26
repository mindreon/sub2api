ALTER TABLE channel_wallet_transactions
    DROP CONSTRAINT IF EXISTS channel_wallet_transactions_type_check;

ALTER TABLE channel_wallet_transactions
    ADD CONSTRAINT channel_wallet_transactions_type_check CHECK (
        transaction_type IN (
            'recharge',
            'refund',
            'consume',
            'commission_reserve',
            'commission_release',
            'commission_settle',
            'commission_deduct',
            'commission_refund'
        )
    );
