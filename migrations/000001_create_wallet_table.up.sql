CREATE TABLE wallets (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    balance DECIMAL(12, 2) NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
