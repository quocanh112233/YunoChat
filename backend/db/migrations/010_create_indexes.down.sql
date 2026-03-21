DROP INDEX IF EXISTS idx_notifications_dedup;
DROP INDEX IF EXISTS idx_notifications_unread;
DROP INDEX IF EXISTS idx_notifications_recipient;

DROP INDEX IF EXISTS idx_attachments_message;

DROP INDEX IF EXISTS idx_messages_sender;
DROP INDEX IF EXISTS idx_messages_created;
DROP INDEX IF EXISTS idx_messages_conversation_cursor;

DROP INDEX IF EXISTS idx_cp_last_read;
DROP INDEX IF EXISTS idx_cp_user;
DROP INDEX IF EXISTS idx_cp_conversation;

DROP INDEX IF EXISTS idx_conversations_dm_lookup;
DROP INDEX IF EXISTS idx_conversations_type;
DROP INDEX IF EXISTS idx_conversations_activity;

DROP INDEX IF EXISTS idx_friendships_addressee;
DROP INDEX IF EXISTS idx_friendships_requester;
DROP INDEX IF EXISTS idx_friendships_canonical;

DROP INDEX IF EXISTS idx_refresh_tokens_expires;
DROP INDEX IF EXISTS idx_refresh_tokens_user;
DROP INDEX IF EXISTS idx_refresh_tokens_hash;

DROP INDEX IF EXISTS idx_users_username_trgm;
DROP INDEX IF EXISTS idx_users_display;
DROP INDEX IF EXISTS idx_users_username;
DROP INDEX IF EXISTS idx_users_email;
