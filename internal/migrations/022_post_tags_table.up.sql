WITH RECURSIVE split(post_id, uid, rest) AS (
    SELECT id, '', tags || ':'
    FROM posts
    WHERE tags IS NOT NULL AND tags != ''
    UNION ALL
    SELECT
        post_id,
        trim(substr(rest, 1, instr(rest, ':') - 1)),
        substr(rest, instr(rest, ':') + 1)
    FROM split
    WHERE rest != ''
)
INSERT OR IGNORE INTO post_user_tags (user_id, post_id)
SELECT CAST(uid AS INTEGER), post_id
FROM split
WHERE uid != ''
    AND CAST(uid AS INTEGER) > 0
    AND EXISTS (SELECT 1 FROM user WHERE user.id = CAST(split.uid AS INTEGER));

WITH RECURSIVE split(post_id, uid, rest) AS (
    SELECT id, '', tags || ':'
    FROM group_posts
    WHERE tags IS NOT NULL AND tags != ''
    UNION ALL
    SELECT
        post_id,
        trim(substr(rest, 1, instr(rest, ':') - 1)),
        substr(rest, instr(rest, ':') + 1)
    FROM split
    WHERE rest != ''
)
INSERT OR IGNORE INTO group_post_user_tags (user_id, post_id)
SELECT CAST(uid AS INTEGER), post_id
FROM split
WHERE uid != ''
    AND CAST(uid AS INTEGER) > 0
    AND EXISTS (SELECT 1 FROM user WHERE user.id = CAST(split.uid AS INTEGER));

CREATE INDEX IF NOT EXISTS idx_post_user_tags_post_id ON post_user_tags(post_id);
CREATE INDEX IF NOT EXISTS idx_group_post_user_tags_post_id ON group_post_user_tags(post_id);

ALTER TABLE posts DROP COLUMN tags;
ALTER TABLE group_posts DROP COLUMN tags;
