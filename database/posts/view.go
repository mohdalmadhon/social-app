package posts

import (
	"database/sql"
	"social/database/users"
	"strconv"
	"strings"
)

func CanView(db *sql.DB, userID, postID int) (bool, error) {
	var targetID int
	var private int
	var public int
	var groupID sql.NullInt64

	err := db.QueryRow(`
		SELECT user_id, public, private, group_id
		FROM posts
		WHERE id = ?
	`, postID).Scan(
		&targetID,
		&public,
		&private,
		&groupID,
	)

	if err != nil {
		return false, err
	}

	if public == 1 {
		return true, nil
	}

	if private == 1 {
		isFollower, err := users.IsFollowing(db, userID, targetID)
		if err != nil {
			return false, err
		}

		return isFollower, nil
	}

	if groupID.Valid {
		var usersArray string

		err := db.QueryRow(`
			SELECT users
			FROM user_posts_groups
			WHERE id = ?
		`, groupID.Int64).Scan(&usersArray)

		if err != nil {
			return false, err
		}

		usersList := strings.Split(usersArray, ":")

		userIDString := strconv.Itoa(userID)

		for _, id := range usersList {
			if id == userIDString {
				return true, nil
			}
		}

		return false, nil
	}

	return false, nil
}
