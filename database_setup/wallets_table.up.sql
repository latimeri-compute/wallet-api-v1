CREATE TABLE IF NOT EXISTS public.wallets (
    id uuid PRIMARY KEY NOT NULL,
    balance numeric(20,8) DEFAULT 0 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
);

INSERT INTO wallets (id, balance)
VALUES
('81a4c5c8-0085-45c1-9c44-d05912276715', 1000.00000000),
('a10c1759-ba1a-47a9-86f7-de80387fc3d4', 32.67000000)
ON CONFLICT (id) DO NOTHING;