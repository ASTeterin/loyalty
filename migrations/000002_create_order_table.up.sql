CREATE TABLE public.orders (
      id SERIAL PRIMARY KEY NOT NULL,
      user_id UUID NOT NULL,
      status VARCHAR(255) NOT NULL,
      created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);