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

	if mode == "none" {
		return false, nil
	}

	inviterFollows, err := users.IsFollowing(db, inviterID, targetID)
	if err != nil {
		return false, err
	}

	targetFollows, err := users.IsFollowing(db, targetID, inviterID)
	if err != nil {
		return false, err
	}

	switch mode {
	case "friends":
		return inviterFollows && targetFollows, nil

	case "following":
		return inviterFollows || targetFollows, nil
	}

	return false, nil
}