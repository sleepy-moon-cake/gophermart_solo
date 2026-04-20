CREATE TABLE IF NOT EXISTS balance (
    id SERIAL PRIMARY KEY,
    owner_id INTEGER NOT NULL,
    current DECIMAL(15, 2) NOT NULL DEFAULT 0,
    withdrawn DECIMAL(15, 2) NOT NULL DEFAULT 0,
    
    CONSTRAINT fk_user
        FOREIGN KEY(owner_id) 
        REFERENCES users(id)
        ON DELETE CASCADE
);