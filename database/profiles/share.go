package profiles

import (
	"database/sql"
	"social/database/chats"
	"social/internal/models"
)

const maxShareUsers = 30

func SearchShareProfile(db *sql.DB, userID int, search string) ([]models.UserRegistration, error) {
	var candidates []models.UserRegistration

	query := `
		SELECT DISTINCT
			u.id,
			COALESCE(u.first_name, ''),
			COALESCE(u.last_name, ''),
			COALESCE(u.username, ''),
			COALESCE(p.avatar_path, '')
		FROM user u
		LEFT JOIN profile p ON p.user_id = u.id
		WHERE u.id <> ?
		AND (
			EXISTS (
				SELECT 1
				FROM user_followers uf
				WHERE uf.follower_id = ?
				AND uf.target_id = u.id
				AND uf.status = 1
			)
			OR EXISTS (
				SELECT 1
				FROM user_followers uf1
				JOIN user_followers uf2
					ON uf1.follower_id = uf2.target_id
					AND uf1.target_id = uf2.follower_id
				WHERE uf1.follower_id = ?
				AND uf1.target_id = u.id
				AND uf1.status = 1
				AND uf2.status = 1
			)
			OR EXISTS (
				SELECT 1
				FROM groups_users gu
				JOIN groups g ON g.id = gu.group_id
				JOIN groups_users gu2 ON gu2.group_id = gu.group_id
				WHERE g.is_private_chat = 1
				AND gu.user_id = u.id
				AND COALESCE(gu.status, 1) = 1
				AND gu2.user_id = ?
				AND COALESCE(gu2.status, 1) = 1
			)
		)
	`

	args := []interface{}{userID, userID, userID, userID}

	if search != "" {
		query += `
			AND (
				u.username LIKE ?
				OR u.first_name LIKE ?
				OR u.last_name LIKE ?
				OR (u.first_name || ' ' || u.last_name) LIKE ?
			)
		`

		pattern := "%" + search + "%"
		args = append(args, pattern, pattern, pattern, pattern)
	}

	query += `
		ORDER BY u.first_name, u.last_name
	`

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user models.UserRegistration

		if err := rows.Scan(
			&user.ID,
			&user.FirstName,
			&user.LastName,
			&user.UserName,
			&user.Avatar,
		); err != nil {
			return nil, err
		}

		candidates = append(candidates, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	users := make([]models.UserRegistration, 0, maxShareUsers)

	for _, user := range candidates {
		canSend, err := chats.CanSendMessage(db, userID, user.ID)
		if err != nil {
			return nil, err
		}

		if !canSend {
			continue
		}

		users = append(users, user)

		if len(users) >= maxShareUsers {
			break
		}
	}

	return users, nil
}

func GetShareProfileCard(db *sql.DB, profileID int) (models.UserRegistration, error) {
	var user models.UserRegistration

	err := db.QueryRow(`
		SELECT
			u.id,
			COALESCE(u.first_name, ''),
			COALESCE(u.last_name, ''),
			COALESCE(u.username, ''),
			COALESCE(p.avatar_path, '')
		FROM user u
		LEFT JOIN profile p ON p.user_id = u.id
		WHERE u.id = ?
	`, profileID).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.UserName,
		&user.Avatar,
	)

	return user, err
}
