CREATE TABLE messages (
    id              UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID            NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    sender_id       UUID            NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    body            TEXT,
    type            VARCHAR(12)     NOT NULL DEFAULT 'TEXT' CHECK (type IN ('TEXT', 'ATTACHMENT')),
    status          VARCHAR(10)     NOT NULL DEFAULT 'SENT' CHECK (status IN ('SENT', 'DELIVERED', 'READ')),
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ,
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_message_content CHECK (
        body IS NOT NULL OR type = 'ATTACHMENT'
    )
);
