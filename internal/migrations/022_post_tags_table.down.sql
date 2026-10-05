ALTER TABLE posts ADD COLUMN tags TEXT;
ALTER TABLE group_posts ADD COLUMN tags TEXT;

UPDATE posts
SET tags = (
    SELECT group_concat(user_id, ':')
    FROM post_user_tags
    WHERE post_user_tags.post_id = posts.id
);

UPDATE group_posts
SET tags = (
    SELECT group_concat(user_id, ':')
    FROM group_post_user_tags
    WHERE group_post_user_tags.post_id = group_posts.id
);

DROP INDEX IF EXISTS idx_post_user_tags_post_id;
DROP INDEX IF EXISTS idx_group_post_user_tags_post_id;
