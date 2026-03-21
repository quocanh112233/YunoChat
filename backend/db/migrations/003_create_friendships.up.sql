CREATE TABLE friendships (
    id              UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    requester_id    UUID            NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    addressee_id    UUID            NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status          VARCHAR(10)     NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'ACCEPTED')),
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_no_self_friend   CHECK (requester_id <> addressee_id)
);
