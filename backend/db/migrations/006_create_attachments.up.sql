CREATE TABLE attachments (
    id              UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id      UUID            NOT NULL UNIQUE REFERENCES messages(id) ON DELETE CASCADE,
    storage_type    VARCHAR(10)     NOT NULL CHECK (storage_type IN ('CLOUDINARY', 'R2')),
    file_type       VARCHAR(5)      NOT NULL CHECK (file_type IN ('IMAGE', 'VIDEO', 'FILE')),
    url             TEXT            NOT NULL,
    thumbnail_url   TEXT,
    original_name   VARCHAR(255)    NOT NULL,
    mime_type       VARCHAR(100)    NOT NULL,
    size_bytes      BIGINT          NOT NULL,
    width           INTEGER,
    height          INTEGER,
    duration_secs   INTEGER,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);
