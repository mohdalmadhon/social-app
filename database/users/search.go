package users

import (
	"database/sql"

	"social/database/dbutil"
	"social/internal/models"
)

func SearchUsers(db *sql.DB, userID int, search string, limit, offset int) ([]models.SearchUser, error) {
	pattern := dbutil.LikePattern(search)

	rows, err := db.Query(`
		SELECT
			u.id,
			u.first_name,
			u.last_name,
			u.username,
			p.avatar_path,
			COALESCE(p.is_private, 0),
			COALESCE((
				SELECT uf.status
				FROM user_followers uf
				WHERE uf.follower_id = ?
					AND uf.target_id = u.id
			), -1) AS follow_status,
			EXISTS (
				SELECT 1
				FROM user_followers back
				WHERE back.follower_id = u.id
					AND back.target_id = ?
					AND back.status = 1
			) AS follows_me
		FROM user u
		LEFT JOIN profile p
			ON p.user_id = u.id
		WHERE u.username LIKE ? ESCAPE '\'
			OR u.first_name LIKE ? ESCAPE '\'
			OR u.last_name LIKE ? ESCAPE '\'
			OR (u.first_name || ' ' || u.last_name) LIKE ? ESCAPE '\'
		ORDER BY
			CASE WHEN follow_status = 1 THEN 0 ELSE 1 END,
			u.first_name COLLATE NOCASE,
			u.last_name COLLATE NOCASE,
			u.id
		LIMIT ? OFFSET ?
	`, userID, userID, pattern, pattern, pattern, pattern, limit, offset)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	result := make([]models.SearchUser, 0)

	for rows.Next() {
		var u models.SearchUser
		var username sql.NullString
		var avatar sql.NullString
		var isPrivate int
		var followsMe bool

		if err := rows.Scan(
			&u.ID,
			&u.FirstName,
			&u.LastName,
			&username,
			&avatar,
			&isPrivate,
			&u.FollowStatus,
			&followsMe,
		); err != nil {
			return nil, err
		}

		u.Username = username.String
		u.Avatar = avatar.String
		u.IsPrivate = isPrivate == 1
		u.IsFriend = u.FollowStatus == 1 && followsMe
		u.IsMe = u.ID == userID

		result = append(result, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
