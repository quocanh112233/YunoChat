-- users
CREATE UNIQUE INDEX idx_users_email    ON users (LOWER(email));
CREATE UNIQUE INDEX idx_users_username ON users (LOWER(username));
CREATE INDEX        idx_users_display  ON users USING GIN (display_name gin_trgm_ops);
CREATE INDEX        idx_users_username_trgm ON users USING GIN (username gin_trgm_ops);

-- refresh_tokens
CREATE UNIQUE INDEX idx_refresh_tokens_hash     ON refresh_tokens (token_hash);
CREATE INDEX        idx_refresh_tokens_user     ON refresh_tokens (user_id);
CREATE INDEX        idx_refresh_tokens_expires  ON refresh_tokens (expires_at) WHERE is_revoked = FALSE;

-- friendships
CREATE UNIQUE INDEX idx_friendships_canonical
    ON friendships (
        LEAST(requester_id::text, addressee_id::text),
        GREATEST(requester_id::text, addressee_id::text)
    );
CREATE INDEX idx_friendships_requester ON friendships (requester_id, status);
CREATE INDEX idx_friendships_addressee ON friendships (addressee_id, status);

-- conversations
CREATE INDEX idx_conversations_activity ON conversations (last_activity_at DESC);
CREATE INDEX idx_conversations_type     ON conversations (type);
CREATE INDEX idx_conversations_dm_lookup ON conversations (type) WHERE type = 'DM';

-- conversation_participants
CREATE INDEX idx_cp_conversation  ON conversation_participants (conversation_id) WHERE left_at IS NULL;
CREATE INDEX idx_cp_user          ON conversation_participants (user_id) WHERE left_at IS NULL;
CREATE INDEX idx_cp_last_read     ON conversation_participants (last_read_message_id);

-- messages
CREATE INDEX idx_messages_conversation_cursor ON messages (conversation_id, created_at DESC, id DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_messages_created ON messages (created_at DESC);
CREATE INDEX idx_messages_sender ON messages (sender_id, conversation_id);

-- attachments
CREATE INDEX idx_attachments_message ON attachments (message_id);

-- notifications
CREATE INDEX idx_notifications_recipient ON notifications (recipient_id, created_at DESC);
CREATE INDEX idx_notifications_unread ON notifications (recipient_id) WHERE is_read = FALSE;
CREATE UNIQUE INDEX idx_notifications_dedup ON notifications (recipient_id, type, reference_id);
