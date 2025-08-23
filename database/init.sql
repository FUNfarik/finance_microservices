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


INSERT INTO users (username, email, password_hash, cash) VALUES
    ('testuser', 'test@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMye1VQ2a/tT/O7xVL3OoYhZLx6BZ7nJxO2', 500.00);


INSERT INTO stock_prices (symbol, name, price, change_percent) VALUES
    ('AAPL', 'Apple Inc.', 175.25, 1.2),
    ('GOOGL', 'Alphabet Inc.', 2750.80, -0.5),
    ('MSFT', 'Microsoft Corporation', 338.50, 0.8),
    ('SPY', 'SPDR S&P 500 ETF', 445.60, 0.3);

-- 3. Add transaction history (showing how testuser spent $7,500 from original $10,000)
INSERT INTO transactions (user_id, symbol, shares, price, transaction_type, total_amount, created_at) VALUES
    (1, 'AAPL', 10, 170.00, 'BUY', 1700.00, NOW() - INTERVAL '7 days'),
    (1, 'GOOGL', 2, 2800.00, 'BUY', 5600.00, NOW() - INTERVAL '5 days'),
    (1, 'SPY', 5, 440.00, 'BUY', 2200.00, NOW() - INTERVAL '2 days');
-- Total spent: $9,500, remaining cash: $500

-- 4. Add current holdings (must match transaction totals)
INSERT INTO holdings (user_id, symbol, shares, avg_price) VALUES
    (1, 'AAPL', 10, 170.00),
    (1, 'GOOGL', 2, 2800.00),
    (1, 'SPY', 5, 440.00);