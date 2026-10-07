ALTER TABLE groups_users
ADD COLUMN last_read_message_id INTEGER NOT NULL DEFAULT 0;

-- existing history counts as already read, only new messages are unread
UPDATE groups_users
SET last_read_message_id = COALESCE(
    (SELECT MAX(m.id) FROM messages m WHERE m.group_id = groups_users.group_id),
    0
);
