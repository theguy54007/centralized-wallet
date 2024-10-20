-- Down Migration: Add from_user_id and to_user_id back, remove from_wallet_number and to_wallet_number
ALTER TABLE transactions
ADD COLUMN from_user_id INT,
ADD COLUMN to_user_id INT;

ALTER TABLE transactions
DROP COLUMN from_wallet_number,
DROP COLUMN to_wallet_number;
