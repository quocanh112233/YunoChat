CREATE TABLE conversations (
    id                   UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    type                 VARCHAR(5)      NOT NULL CHECK (type IN ('DM', 'GROUP')),
    name                 VARCHAR(100),
    avatar_url           TEXT,
    avatar_cloudinary_id TEXT,
    last_message_id      UUID,
    last_activity_at     TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    created_at           TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ     NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_group_name CHECK (
        type = 'DM' OR (type = 'GROUP' AND name IS NOT NULL)
    )
);
