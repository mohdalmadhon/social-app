-- group id refrence to if the post is 
--  public: group id == null
CREATE TABLE IF NOT EXISTS posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    content VARCHAR(1000) NOT NULL,
    image_path TEXT,
    allow_comments INTEGER NOT NULL DEFAULT 1,
    location TEXT,
    group_id INTEGER,
    public int CHECK (public in (1,0)),
    private int CHECK (public in (1,0)),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    like_count INTEGER NOT NULL DEFAULT 0,
    dislike_count INTEGER NOT NULL DEFAULT 0,
    comment_count INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY (user_id)
        REFERENCES user(id)
        ON DELETE CASCADE,

    FOREIGN KEY (group_id)
        REFERENCES user_posts_groups(id)
        ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS user_posts_groups (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(15) NOT NULL,
    user_id INTEGER NOT NULL,
    users TEXT NOT NULL DEFAULT '',
    UNIQUE (user_id, name),

    FOREIGN KEY (user_id)
        REFERENCES user(id)
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS post_user_tags (
    user_id INTEGER NOT NULL,
    post_id INTEGER NOT NULL,
    PRIMARY KEY (user_id, post_id),
    FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE,
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE
);


CREATE TRIGGER trg_increase_post
AFTER INSERT ON posts
FOR EACH ROW
BEGIN
    UPDATE profile
    SET num_of_posts = num_of_posts + 1
    WHERE user_id = NEW.user_id;
END;

DROP TRIGGER IF EXISTS trg_decrease_post;

CREATE TRIGGER trg_decrease_post
AFTER DELETE ON posts
FOR EACH ROW
BEGIN
    UPDATE profile
    SET num_of_posts = num_of_posts - 1
    WHERE user_id = OLD.user_id;
END;
