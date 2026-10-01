CREATE TABLE IF NOT EXISTS user_preferences (
    user_ID PRIMARY KEY,
    chat TEXT NOT NULL CHECK (
        chat IN (
            'any',
            'following',
            'following-followers',
            'friends',
            'friends-following',
            'none'
        )
    ) DEFAULT 'following-followers',
    FOREIGN KEY (user_ID) REFERENCES user(id) ON DELETE CASCADE
);

CREATE TRIGGER IF NOT EXISTS trg_user_preferences
AFTER INSERT ON user
FOR EACH ROW
BEGIN
    INSERT INTO user_preferences (user_ID)
    VALUES (NEW.id);
END;