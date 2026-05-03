CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    owner_id INTEGER NOT NULL,
    number VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'NEW',
    accrual INTEGER NOT NULL DEFAULT 0,
    uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Уникальность номера (для кодов 200 и 409)
    UNIQUE (number),
    
    -- Связь с таблицей пользователей (Foreign Key)
    CONSTRAINT fk_user
        FOREIGN KEY(owner_id) 
        REFERENCES users(id)
        ON DELETE CASCADE
);