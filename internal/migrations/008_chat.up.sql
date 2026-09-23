CREATE TABLE IF NOT EXISTS groups (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    is_private_chat INTEGER NOT NULL CHECK (is_private_chat IN (1, 0)),
    owner_id INTEGER,
    name VARCHAR(15),
    description VARCHAR(200),
    avatar TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (owner_id) REFERENCES user(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS groups_users (
    group_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    status INTEGER CHECK (status IN (1,0)),
    PRIMARY KEY (group_id, user_id),

    FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE,
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sender_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    group_id INTEGER NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (sender_id) REFERENCES user(id) ON DELETE CASCADE,
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE
);

DROP TRIGGER IF EXISTS trg_groups_users_before;

CREATE TRIGGER trg_groups_users_before
BEFORE INSERT ON groups_users
FOR EACH ROW
BEGIN
    SELECT CASE
        WHEN
            (SELECT is_private_chat
             FROM groups
             WHERE id = NEW.group_id) = 1
            AND
            (SELECT COUNT(*)
             FROM groups_users
             WHERE group_id = NEW.group_id) >= 2
        THEN RAISE(ABORT, 'Private chat already has two users')

        WHEN EXISTS (
            SELECT 1
            FROM groups_users
            WHERE user_id = NEW.user_id
            AND group_id = NEW.group_id
        )
        THEN RAISE(ABORT, 'Member already exists in this chat')
    END;
END;

DROP TRIGGER IF EXISTS trg_user_group_exists;

CREATE TRIGGER trg_user_group_exists
BEFORE INSERT ON messages
FOR EACH ROW
BEGIN
    SELECT CASE
        WHEN NOT EXISTS (
            SELECT 1
            FROM groups_users
            WHERE user_id = NEW.sender_id
            AND group_id = NEW.group_id
        )
        THEN RAISE(ABORT, 'user does not belong to this chat')
    END;
END;

