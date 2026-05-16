CREATE TABLE IF NOT EXISTS withdrawals (
    id SERIAL PRIMARY KEY,
    owner_id INTEGER NOT NULL,
    order_number VARCHAR(255) NOT NULL,
    sum INTEGER NOT NULL DEFAULT 0,
    processed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(order_number),

    CONSTRAINT fk_withdrawals_user 
        FOREIGN KEY(owner_id) 
        REFERENCES users(id) 
        ON DELETE CASCADE
)
