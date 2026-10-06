UPDATE groups
SET avatar = 'groups/avatars/' || substr(avatar, 7)
WHERE avatar LIKE 'posts/%';
