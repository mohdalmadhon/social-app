DROP TRIGGER IF EXISTS trg_group_post_reaction_delete;
DROP TRIGGER IF EXISTS trg_group_post_reaction_update;
DROP TRIGGER IF EXISTS trg_group_post_reaction_insert;
DROP TRIGGER IF EXISTS trg_comment_count_delete;
DROP TRIGGER IF EXISTS trg_group_comment_count_insert;

DROP INDEX IF EXISTS idx_group_comment_votes_user_id;
DROP INDEX IF EXISTS idx_group_comments_reply_to;
DROP INDEX IF EXISTS idx_group_comments_user_id;
DROP INDEX IF EXISTS idx_group_comments_post_id;
DROP INDEX IF EXISTS idx_post_reactions_user_id;

DROP TABLE IF EXISTS group_comment_votes;
DROP TABLE IF EXISTS group_comments;
DROP TABLE IF EXISTS group_post_reactions;
DROP TABLE IF EXISTS group_post_user_tags;
DROP TABLE IF EXISTS group_posts;