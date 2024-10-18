CREATE TABLE transactions (
  id SERIAL PRIMARY KEY,
  user_id INT NOT NULL,
  transaction_type VARCHAR(50) NOT NULL, -- "deposit" or "withdraw"
  amount DECIMAL(10, 2) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

  -- Foreign key constraint to link transactions to a specific user
  FOREIGN KEY (user_id) REFERENCES users(id)
);
