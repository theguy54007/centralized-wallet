CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    from_wallet_number VARCHAR(50),
    to_wallet_number VARCHAR(50),
    transaction_type VARCHAR(50) NOT NULL,
    amount NUMERIC(12, 2) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Add indexes on from_wallet_number and to_wallet_number columns
CREATE INDEX idx_from_wallet_number ON transactions(from_wallet_number);
CREATE INDEX idx_to_wallet_number ON transactions(to_wallet_number);
