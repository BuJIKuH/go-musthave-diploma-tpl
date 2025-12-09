CREATE TABLE IF NOT EXISTS users (
                       id UUID PRIMARY KEY,
                       login TEXT NOT NULL UNIQUE,
                       password_hash TEXT NOT NULL,
                       created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE orders (
                        id UUID PRIMARY KEY,
                        number TEXT NOT NULL UNIQUE,
                        user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                        status TEXT NOT NULL DEFAULT 'NEW',
                        accrual DECIMAL(12,2),
                        uploaded_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE withdrawals (
                             id UUID PRIMARY KEY,
                             user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                             order_number TEXT NOT NULL,
                             sum DECIMAL(12,2) NOT NULL,
                             processed_at TIMESTAMPTZ DEFAULT NOW(),
                             UNIQUE(user_id, order_number)
);