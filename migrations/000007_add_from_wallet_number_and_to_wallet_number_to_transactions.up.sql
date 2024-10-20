-- Up Migration: Add from_wallet_number and to_wallet_number, remove from_user_id and to_user_id
ALTER TABLE transactions
ADD COLUMN from_wallet_number VARCHAR(255),
ADD COLUMN to_wallet_number VARCHAR(255);

ALTER TABLE transactions
DROP COLUMN from_user_id,
DROP COLUMN to_user_id;
