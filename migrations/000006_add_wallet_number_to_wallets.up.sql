ALTER TABLE wallets
ADD COLUMN wallet_number VARCHAR(255) UNIQUE NOT NULL;

-- Create an index on the wallet_number column
CREATE INDEX idx_wallet_number ON wallets(wallet_number);
