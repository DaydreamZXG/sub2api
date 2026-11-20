-- Resellers hand out keys to their own clients at their own price.
-- A client key counts what a call costs its owner, times client_markup, against the key's quota:
-- the client's budget is in the client's price. 0 for an ordinary key.
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS client_markup DECIMAL(10,4) NOT NULL DEFAULT 0;

-- Who may create client keys; an admin adds and removes them.
CREATE TABLE IF NOT EXISTS resellers (
    user_id    BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
