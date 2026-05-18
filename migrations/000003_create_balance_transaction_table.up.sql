CREATE TABLE public.balance_transaction (
    order_id VARCHAR(20) PRIMARY KEY NOT NULL,
    user_id UUID NOT NULL,
    points DECIMAL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);