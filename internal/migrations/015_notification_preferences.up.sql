CREATE TABLE IF NOT EXISTS notification_preferences (
    user_id INTEGER NOT NULL,
    type TEXT NOT NULL CHECK (
        type IN (
            'follow',
            'post_reaction',
            'comment',
            'comment_like',
            'mention',
            'event',
            'message'
        )
    ),
    mode TEXT NOT NULL CHECK (mode IN ('any', 'friends', 'none')) DEFAULT 'any',
    PRIMARY KEY (user_id, type),
    FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE
);
