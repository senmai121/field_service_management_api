-- Migrate auth users from public.users into the fsm schema.
-- After running this migration, the API will use fsm.users exclusively.

CREATE TABLE IF NOT EXISTS fsm.users (
    id          bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    username    text        NOT NULL,
    email       text        NOT NULL UNIQUE,
    password    text        NOT NULL,
    system_id   bigint,
    created_at  timestamptz NOT NULL DEFAULT now()
);

-- Copy existing users (skip conflicts in case migration is re-run).
INSERT INTO fsm.users (username, email, password, system_id, created_at)
SELECT username, email, password, system_id, created_at
FROM   public.users
ON CONFLICT (email) DO NOTHING;
