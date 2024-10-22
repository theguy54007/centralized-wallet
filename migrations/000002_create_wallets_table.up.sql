CREATE TABLE IF NOT EXISTS wallets (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    wallet_number VARCHAR(50) NOT NULL UNIQUE,
    balance NUMERIC(12, 2) NOT NULL DEFAULT 0.00,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Add an index on the wallet_number column
CREATE INDEX idx_wallet_number ON wallets(wallet_number);

-- Add an index on the user_id column
CREATE INDEX idx_wallet_user_id ON wallets(user_id);
