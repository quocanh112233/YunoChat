CREATE TABLE notifications (
    id              UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    recipient_id    UUID            NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    actor_id        UUID            NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type            VARCHAR(20)     NOT NULL CHECK (type IN (
                                        'FRIEND_REQUEST',
                                        'FRIEND_ACCEPTED',
                                        'GROUP_ADDED'
                                    )),
    reference_id    UUID            NOT NULL,
    reference_type  VARCHAR(15)     NOT NULL CHECK (reference_type IN ('friendship', 'conversation')),
    is_read         BOOLEAN         NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    read_at         TIMESTAMPTZ,

    CONSTRAINT chk_read_consistency CHECK (
        (is_read = FALSE AND read_at IS NULL) OR
        (is_read = TRUE  AND read_at IS NOT NULL)
    )
);
