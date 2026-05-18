CREATE TABLE public.balance_transaction (
    id SERIAL PRIMARY KEY NOT NULL,
    order_id INTEGER NOT NULL,
    user_id UUID NOT NULL,
    points DECIMAL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);