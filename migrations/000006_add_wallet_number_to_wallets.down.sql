-- Drop the index on the wallet_number column
DROP INDEX IF EXISTS idx_wallet_number;

-- Remove the wallet_number column from the wallets table
ALTER TABLE wallets
DROP COLUMN IF EXISTS wallet_number;
