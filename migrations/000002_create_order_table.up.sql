CREATE TABLE public.orders (
      id VARCHAR(20) PRIMARY KEY NOT NULL,
      user_id UUID NOT NULL,
      status VARCHAR(255) NOT NULL,
      created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);