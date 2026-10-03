ALTER TABLE user_preferences
ADD COLUMN group_invite TEXT NOT NULL DEFAULT 'following' CHECK (group_invite IN ('friends', 'following', 'none'));
