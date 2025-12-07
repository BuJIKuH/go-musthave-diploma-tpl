CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
                       id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                       login TEXT NOT NULL UNIQUE,
                       password_hash TEXT NOT NULL,
                       created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE orders (
                        id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                        number TEXT NOT NULL UNIQUE,
                        user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                        status TEXT NOT NULL DEFAULT 'NEW',
                        accrual DECIMAL(12,2),
                        uploaded_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE withdrawals (
                             id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                             user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                             order_number TEXT NOT NULL,
                             sum DECIMAL(12,2) NOT NULL,
                             processed_at TIMESTAMPTZ DEFAULT NOW(),
                             UNIQUE(user_id, order_number)
);