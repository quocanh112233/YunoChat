CREATE TABLE users (
    id                      UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    email                   VARCHAR(255)    NOT NULL,
    username                VARCHAR(30)     NOT NULL,
    password_hash           VARCHAR(255)    NOT NULL,
    display_name            VARCHAR(50)     NOT NULL,
    bio                     VARCHAR(160),
    avatar_url              TEXT,
    avatar_cloudinary_id    TEXT,
    status                  VARCHAR(10)     NOT NULL DEFAULT 'OFFLINE' CHECK (status IN ('ONLINE', 'OFFLINE')),
    last_seen_at            TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    created_at              TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);
