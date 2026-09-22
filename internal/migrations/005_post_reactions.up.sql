

CREATE TABLE IF NOT EXISTS post_reactions (
    post_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    value INTEGER NOT NULL CHECK (value IN (1, -1)),
    PRIMARY KEY (post_id, user_id),
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS comments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    content VARCHAR(200) not null,
    user_id INTEGER not null,
    post_id INTEGER not null,
    reply_to INTEGER,
    votes INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE,
    FOREIGN KEY (reply_to) REFERENCES comments(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS comment_votes (
    user_id INTEGER NOT NULL,
    comment_id INTEGER NOT NULL,
    count INTEGER NOT NULL CHECK (count IN (1, -1)),
    PRIMARY KEY (comment_id, user_id),
    FOREIGN KEY (comment_id) REFERENCES comments(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE
);

CREATE INDEX idx_post_reactions_user_id
ON post_reactions(user_id);

CREATE INDEX IF NOT EXISTS idx_comments_post_id
ON comments(post_id);

CREATE INDEX IF NOT EXISTS idx_comments_user_id
ON comments(user_id);

CREATE INDEX IF NOT EXISTS idx_comments_reply_to
ON comments(reply_to);

CREATE INDEX IF NOT EXISTS idx_comment_votes_user_id
ON comment_votes(user_id);

CREATE TRIGGER IF NOT EXISTS trg_comment_count_insert
AFTER INSERT ON comments
FOR EACH ROW
BEGIN
    UPDATE posts
    SET comment_count = comment_count + 1
    WHERE id = NEW.post_id;
END;

CREATE TRIGGER IF NOT EXISTS trg_comment_count_delete
AFTER DELETE ON comments
FOR EACH ROW
BEGIN
    UPDATE posts
    SET comment_count = comment_count - 1
    WHERE id = OLD.post_id;
END;

CREATE TRIGGER IF NOT EXISTS trg_post_reaction_insert
AFTER INSERT ON post_reactions
FOR EACH ROW
BEGIN
    UPDATE posts
    SET
        like_count = like_count + CASE WHEN NEW.value = 1 THEN 1 ELSE 0 END,
        dislike_count = dislike_count + CASE WHEN NEW.value = -1 THEN 1 ELSE 0 END
    WHERE id = NEW.post_id;
END;

CREATE TRIGGER IF NOT EXISTS trg_post_reaction_update
AFTER UPDATE OF value ON post_reactions
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

CREATE TRIGGER IF NOT EXISTS trg_post_reaction_delete
AFTER DELETE ON post_reactions
BEGIN
    UPDATE posts
    SET
        like_count = like_count - CASE WHEN OLD.value = 1 THEN 1 ELSE 0 END,
        dislike_count = dislike_count - CASE WHEN OLD.value = -1 THEN 1 ELSE 0 END
    WHERE id = OLD.post_id;
END;

