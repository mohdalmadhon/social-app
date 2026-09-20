CREATE TABLE IF NOT EXISTS group_posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    content VARCHAR(1000) NOT NULL,
    image_path TEXT,
    location TEXT,
    group_id INTEGER,
    tags TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE,
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE
    SET
        NULL
);

CREATE TABLE IF NOT EXISTS group_post_user_tags (
    user_id INTEGER NOT NULL,
    post_id INTEGER NOT NULL,
    PRIMARY KEY (user_id, post_id),
    FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE,
    FOREIGN KEY (post_id) REFERENCES group_posts(id) ON DELETE CASCADE
);

ALTER TABLE posts ADD COLUMN like_count INTEGER NOT NULL DEFAULT 0;
ALTER TABLE posts ADD COLUMN dislike_count INTEGER NOT NULL DEFAULT 0;
ALTER TABLE posts ADD COLUMN comment_count INTEGER NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS group_post_reactions (
    post_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    value INTEGER NOT NULL CHECK (value IN (1, -1)),
    PRIMARY KEY (post_id, user_id),
    FOREIGN KEY (post_id) REFERENCES group_posts(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS group_comments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    content VARCHAR(200) not null,
    user_id INTEGER not null,
    post_id INTEGER not null,
    reply_to INTEGER,
    votes INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (post_id) REFERENCES group_posts(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE,
    FOREIGN KEY (reply_to) REFERENCES group_comments(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS group_comment_votes (
    user_id INTEGER NOT NULL,
    comment_id INTEGER NOT NULL,
    count INTEGER NOT NULL CHECK (count IN (1, -1)),
    PRIMARY KEY (comment_id, user_id),
    FOREIGN KEY (comment_id) REFERENCES group_comments(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE
);

CREATE INDEX idx_post_reactions_user_id
ON group_post_reactions(user_id);

CREATE INDEX IF NOT EXISTS idx_group_comments_post_id
ON group_comments(post_id);

CREATE INDEX IF NOT EXISTS idx_group_comments_user_id
ON group_comments(user_id);

CREATE INDEX IF NOT EXISTS idx_group_comments_reply_to
ON group_comments(reply_to);

CREATE INDEX IF NOT EXISTS idx_group_comment_votes_user_id
ON group_comment_votes(user_id);

CREATE TRIGGER IF NOT EXISTS trg_group_comment_count_insert
AFTER INSERT ON group_comments
FOR EACH ROW
BEGIN
    UPDATE posts
    SET comment_count = comment_count + 1
    WHERE id = NEW.post_id;
END;

CREATE TRIGGER IF NOT EXISTS trg_comment_count_delete
AFTER DELETE ON group_comments
FOR EACH ROW
BEGIN
    UPDATE posts
    SET comment_count = comment_count - 1
    WHERE id = OLD.post_id;
END;

CREATE TRIGGER IF NOT EXISTS trg_group_post_reaction_insert
AFTER INSERT ON group_post_reactions
FOR EACH ROW
BEGIN
    UPDATE posts
    SET
        like_count = like_count + CASE WHEN NEW.value = 1 THEN 1 ELSE 0 END,
        dislike_count = dislike_count + CASE WHEN NEW.value = -1 THEN 1 ELSE 0 END
    WHERE id = NEW.post_id;
END;

CREATE TRIGGER IF NOT EXISTS trg_group_post_reaction_update
AFTER UPDATE OF value ON group_post_reactions
WHEN OLD.value != NEW.value
BEGIN
    UPDATE posts
    SET
        like_count = like_count
            - CASE WHEN OLD.value = 1 THEN 1 ELSE 0 END
            + CASE WHEN NEW.value = 1 THEN 1 ELSE 0 END,
        dislike_count = dislike_count
            - CASE WHEN OLD.value = -1 THEN 1 ELSE 0 END
            + CASE WHEN NEW.value = -1 THEN 1 ELSE 0 END
    WHERE id = NEW.post_id;
END;

CREATE TRIGGER IF NOT EXISTS trg_group_post_reaction_delete
AFTER DELETE ON group_post_reactions
BEGIN
    UPDATE posts
    SET
        like_count = like_count - CASE WHEN OLD.value = 1 THEN 1 ELSE 0 END,
        dislike_count = dislike_count - CASE WHEN OLD.value = -1 THEN 1 ELSE 0 END
    WHERE id = OLD.post_id;
END;

