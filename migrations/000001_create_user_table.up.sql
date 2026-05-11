CREATE TABLE public.users (
     id UUID PRIMARY KEY NOT NULL,
     login VARCHAR(255) NOT NULL,
     pass_hash VARCHAR(255) NOT NULL
);