package preferences

import (
	"database/sql"

	"social/database/users"
)

func CanInviteToGroup(db *sql.DB, inviterID, targetID int) (bool, error) {
	var mode string

	err := db.QueryRow(`
		SELECT group_invite
		FROM user_preferences
		WHERE user_id = ?
	`, targetID).Scan(&mode)

	if err == sql.ErrNoRows {
		mode = "following"
	} else if err != nil {
		return false, err
	}

	switch mode {
	case "none":
		return false, nil

	case "friends":
		return users.IsFriend(db, inviterID, targetID)

	case "following":
		friend, err := users.IsFriend(db, inviterID, targetID)
		if err != nil || friend {
			return friend, err
		}

		return users.IsFollowing(db, targetID, inviterID)
	}

	return false, nil
}
