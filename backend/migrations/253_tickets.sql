-- Support tickets: one row per ticket plus its conversation.
-- author_type separates users, staff and AI so automated replies can be added later.
CREATE TABLE IF NOT EXISTS tickets (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    subject         VARCHAR(200) NOT NULL,
    category        VARCHAR(32) NOT NULL DEFAULT 'other',
    status          VARCHAR(16) NOT NULL DEFAULT 'open',
    last_message_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS ticket_user_id_last_message_at ON tickets (user_id, last_message_at);
CREATE INDEX IF NOT EXISTS ticket_status_last_message_at ON tickets (status, last_message_at);

CREATE TABLE IF NOT EXISTS ticket_messages (
    id          BIGSERIAL PRIMARY KEY,
    ticket_id   BIGINT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    author_type VARCHAR(16) NOT NULL,
    author_id   BIGINT,
    content     TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS ticketmessage_ticket_id_created_at ON ticket_messages (ticket_id, created_at);
