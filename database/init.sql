CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    cash DECIMAL(15,2) DEFAULT 10000.00,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE holdings (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    symbol VARCHAR(10) NOT NULL,
    shares INTEGER NOT NULL,
    avg_price DECIMAL(10,2) NOT NULL,
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, symbol),
    FOREIGN KEY(user_id) REFERENCES users(id)
);

CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    symbol VARCHAR(10) NOT NULL,
    shares INTEGER NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    transaction_type VARCHAR(4) NOT NULL CHECK (transaction_type IN ('BUY', 'SELL')),
    total_amount DECIMAL(15,2) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY(user_id) REFERENCES users(id)
);

CREATE TABLE stock_prices (
    symbol VARCHAR(10) PRIMARY KEY,
    name VARCHAR(255),
    price DECIMAL(10,2) NOT NULL,
    change_percent DECIMAL(5,2),
    updated_at TIMESTAMP DEFAULT NOW()
);