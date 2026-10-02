ALTER TABLE user_preferences
ADD COLUMN show_email TEXT NOT NULL DEFAULT 'any' CHECK (show_email IN ('any', 'none'));

ALTER TABLE user_preferences
ADD COLUMN show_dob TEXT NOT NULL DEFAULT 'any' CHECK (show_dob IN ('any', 'none'));

ALTER TABLE user_preferences
ADD COLUMN additional_info TEXT NOT NULL DEFAULT 'any' CHECK (additional_info IN ('any', 'followers', 'friends'));
