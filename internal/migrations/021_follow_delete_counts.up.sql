DROP TRIGGER IF EXISTS trg_decrease_followers;
DROP TRIGGER IF EXISTS trg_decrease_following;

CREATE TRIGGER trg_decrease_followers
AFTER DELETE ON user_followers
FOR EACH ROW
WHEN OLD.status = 1
BEGIN
    UPDATE profile
    SET num_of_followers = num_of_followers - 1
    WHERE user_id = OLD.target_id;
END;

CREATE TRIGGER trg_decrease_following
AFTER DELETE ON user_followers
FOR EACH ROW
WHEN OLD.status = 1
BEGIN
    UPDATE profile
    SET num_of_following = num_of_following - 1
    WHERE user_id = OLD.follower_id;
END;
