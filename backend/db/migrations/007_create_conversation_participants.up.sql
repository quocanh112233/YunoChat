CREATE TABLE conversation_participants (
    id                      UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id         UUID            NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    user_id                 UUID            NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role                    VARCHAR(10)     NOT NULL DEFAULT 'MEMBER' CHECK (role IN ('MEMBER', 'ADMIN')),
    last_read_message_id    UUID            REFERENCES messages(id) ON DELETE SET NULL,
    last_read_at            TIMESTAMPTZ,
    joined_at               TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    left_at                 TIMESTAMPTZ,

    CONSTRAINT uq_participant UNIQUE (conversation_id, user_id)
);
