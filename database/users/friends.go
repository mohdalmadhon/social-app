package users

import (
	"database/sql"

	"social/internal/models"
)

func GetFriends(db *sql.DB, userID, offset int) (map[int]models.UserRegistration, error) {
	friends := make(map[int]models.UserRegistration)

	rows, err := db.Query(`
		SELECT u.id, u.first_name, u.last_name, p.avatar_path
		FROM user AS u
		LEFT JOIN profile AS p ON u.id = p.user_id
		WHERE EXISTS (
			SELECT 1
			FROM user_followers AS uf1
			WHERE uf1.follower_id = ?
			  AND uf1.target_id = u.id
		)
		AND EXISTS (
			SELECT 1
			FROM user_followers AS uf2
			WHERE uf2.follower_id = u.id
			  AND uf2.target_id = ?
		)
		ORDER BY u.id
		LIMIT 30
		OFFSET ?
	`, userID, userID, offset)

	if err != nil {
		return friends, err
	}
	defer rows.Close()

	for rows.Next() {
		var friend models.UserRegistration
		var id int
		var firstName sql.NullString
		var lastName sql.NullString
		var avatar sql.NullString

		err := rows.Scan(
			&id,
			&firstName,
			&lastName,
			&avatar,
		)

		if err != nil {
			return friends, err
		}

		if firstName.Valid {
			friend.FirstName = firstName.String
		}

		if lastName.Valid {
			friend.LastName = lastName.String
		}

		if avatar.Valid {
			friend.Avatar = avatar.String
		}

		friends[id] = friend
	}

	if err := rows.Err(); err != nil {
		return friends, err
	}

	return friends, nil
}

func SearchFriends(db *sql.DB, userID int, search string) (map[int]models.UserRegistration, error) {
	friends := make(map[int]models.UserRegistration)

	search = "%" + search + "%"

	rows, err := db.Query(`
		SELECT u.id, u.first_name, u.last_name, p.avatar_path
		FROM user AS u
		LEFT JOIN profile AS p ON u.id = p.user_id
		WHERE EXISTS (
			SELECT 1
			FROM user_followers AS uf1
			WHERE uf1.follower_id = ?
			  AND uf1.target_id = u.id
		)
		AND EXISTS (
			SELECT 1
			FROM user_followers AS uf2
			WHERE uf2.follower_id = u.id
			  AND uf2.target_id = ?
		)
		AND (
			u.first_name LIKE ?
			OR u.last_name LIKE ?
		)
		LIMIT 100
	`, userID, userID, search, search)

	if err != nil {
		return friends, err
	}
	defer rows.Close()

	for rows.Next() {
		var user models.UserRegistration
		var id int
		var firstName sql.NullString
		var lastName sql.NullString
		var avatar sql.NullString

		err := rows.Scan(
			&id,
			&firstName,
			&lastName,
			&avatar,
		)

		if err != nil {
			return friends, err
		}

		if firstName.Valid {
			user.FirstName = firstName.String
		}

		if lastName.Valid {
			user.LastName = lastName.String
		}

		if avatar.Valid {
			user.Avatar = avatar.String
		}

		friends[id] = user
	}

	if err := rows.Err(); err != nil {
		return friends, err
	}

	return friends, nil
}

func GetFollowers(db *sql.DB, userID, limit, offset int) ([]models.UserRegistration, error) {
	var followers []models.UserRegistration

	rows, err := db.Query(`
		SELECT u.id, u.first_name, u.last_name, p.avatar_path
		FROM user_followers uf
		JOIN user u ON u.id = uf.follower_id
		LEFT JOIN profile p ON p.user_id = u.id
		WHERE uf.target_id = ?
		ORDER BY u.id
		LIMIT ? OFFSET ?
	`, userID, limit, offset)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user models.UserRegistration
		var firstName sql.NullString
		var lastName sql.NullString
		var avatar sql.NullString

		err := rows.Scan(
			&user.ID,
			&firstName,
			&lastName,
			&avatar,
		)

		if err != nil {
			return nil, err
		}

		if firstName.Valid {
			user.FirstName = firstName.String
		}

		if lastName.Valid {
			user.LastName = lastName.String
		}

		if avatar.Valid {
			user.Avatar = avatar.String
		}

		followers = append(followers, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return followers, nil
}

func GetFollowing(db *sql.DB, userID, limit, offset int) ([]models.UserRegistration, error) {
	var following []models.UserRegistration

	rows, err := db.Query(`
		SELECT u.id, u.first_name, u.last_name, p.avatar_path
		FROM user_followers uf
		JOIN user u ON u.id = uf.target_id
		LEFT JOIN profile p ON p.user_id = u.id
		WHERE uf.follower_id = ?
		ORDER BY u.id
		LIMIT ? OFFSET ?
	`, userID, limit, offset)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user models.UserRegistration
		var firstName sql.NullString
		var lastName sql.NullString
		var avatar sql.NullString

		err := rows.Scan(
			&user.ID,
			&firstName,
			&lastName,
			&avatar,
		)

		if err != nil {
			return nil, err
		}

		if firstName.Valid {
			user.FirstName = firstName.String
		}

		if lastName.Valid {
			user.LastName = lastName.String
		}

		if avatar.Valid {
			user.Avatar = avatar.String
		}

		following = append(following, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return following, nil
}

func IsFriend(db *sql.DB, userID, targetID int) (bool, error) {
	var exists int

	err := db.QueryRow(`
		SELECT 1
		FROM user_followers uf1
		JOIN user_followers uf2
			ON uf1.follower_id = uf2.target_id
			AND uf1.target_id = uf2.follower_id
		WHERE uf1.follower_id = ?
		  AND uf1.target_id = ?
		LIMIT 1
	`, userID, targetID).Scan(&exists)

	if err == sql.ErrNoRows {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

func IsFollowing(db *sql.DB, userID, targetID int) (bool, error) {
	var exists int

	err := db.QueryRow(`
		SELECT 1
		FROM user_followers
		WHERE follower_id = ?
		  AND target_id = ?
		  AND status = 1
		LIMIT 1
	`, userID, targetID).Scan(&exists)

	if err == sql.ErrNoRows {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

func Get_followers_following_chatList(db *sql.DB, userID int, search string) ([]models.UserRegistration, error) {
	var users []models.UserRegistration

	search = "%" + search + "%"

	rows, err := db.Query(`
		SELECT u.id, u.first_name, u.last_name, u.username, p.avatar_path
		FROM user u
		JOIN profile p ON p.user_id = u.id
		WHERE (
			EXISTS (
				SELECT 1
				FROM user_followers
				WHERE follower_id = ? AND target_id = u.id
			)
			OR
			EXISTS (
				SELECT 1
				FROM user_followers
				WHERE target_id = ? AND follower_id = u.id
			)
			OR
			EXISTS (
				SELECT 1
				FROM groups_users gu1
				JOIN groups g ON g.id = gu1.group_id
				JOIN groups_users gu2 ON gu2.group_id = gu1.group_id
				WHERE gu1.user_id = ?
					AND gu2.user_id = u.id
					AND g.is_private_chat = 1
			)
		)
		AND u.id != ?
		AND (
			u.first_name LIKE ?
			OR u.last_name LIKE ?
			OR u.username LIKE ?
		)
		LIMIT 30
	`, userID, userID, userID, userID, search, search, search)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var u models.UserRegistration

		if err := rows.Scan(
			&u.ID,
			&u.FirstName,
			&u.LastName,
			&u.UserName,
			&u.Avatar,
		); err != nil {
			return nil, err
		}

		users = append(users, u)
	}

	return users, rows.Err()
}
