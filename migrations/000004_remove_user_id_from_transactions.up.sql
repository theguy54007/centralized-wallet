-- Up Migration: Remove the "user_id" column from the "transactions" table

ALTER TABLE transactions
DROP COLUMN user_id;
