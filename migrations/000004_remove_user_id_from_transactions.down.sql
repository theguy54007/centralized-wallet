-- Down Migration: Add the "user_id" column back to the "transactions" table

ALTER TABLE transactions
ADD COLUMN user_id INTEGER;
