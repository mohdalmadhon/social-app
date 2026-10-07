ALTER TABLE user_preferences
ADD COLUMN allow_previous_senders INTEGER NOT NULL DEFAULT 0 CHECK (allow_previous_senders IN (0, 1));
