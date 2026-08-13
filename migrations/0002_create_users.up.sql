CREATE TABLE IF NOT EXISTS users (
    id              BIGSERIAL PRIMARY KEY,
    email           TEXT NOT NULL,
    password_hash   TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT users_email_unique UNIQUE (email)
);

-- One-to-one: a user's default search area. Applied by the frontend as
-- the initial filters when they log in.
CREATE TABLE IF NOT EXISTS user_preferences (
    user_id         BIGINT PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    country         TEXT NOT NULL DEFAULT '',
    workplace       TEXT NOT NULL DEFAULT '',
    seniority       TEXT NOT NULL DEFAULT '',
    tag             TEXT NOT NULL DEFAULT '',
    keyword         TEXT NOT NULL DEFAULT '',
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
