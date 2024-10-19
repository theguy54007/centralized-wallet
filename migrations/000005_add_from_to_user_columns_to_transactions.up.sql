-- Add new columns: from_user_id and to_user_id to the transactions table
ALTER TABLE transactions
ADD COLUMN from_user_id INT,
ADD COLUMN to_user_id INT;
