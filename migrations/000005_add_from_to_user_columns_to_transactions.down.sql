-- Drop the columns from_user_id and to_user_id
ALTER TABLE transactions
DROP COLUMN from_user_id,
DROP COLUMN to_user_id;
