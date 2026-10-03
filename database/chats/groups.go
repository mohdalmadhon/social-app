package chats

import (
	"database/sql"
	"social/internal/models"
)

func GetPrivateChatsList(db *sql.DB, userID, offset int) ([]models.PrivateChat, error) {
	chats := make([]models.PrivateChat, 0)

	rows, err := db.Query(`
		SELECT
			u.id,
			u.first_name,
			u.last_name,
			p.avatar_path,
			g.id AS group_id
		FROM groups g
		JOIN groups_users gu1
			ON gu1.group_id = g.id
			AND gu1.user_id = ?
		JOIN groups_users gu2
			ON gu2.group_id = g.id
			AND gu2.user_id != ?
		JOIN user u
			ON u.id = gu2.user_id
		JOIN profile p
			ON p.user_id = u.id
		LEFT JOIN messages m
			ON m.group_id = g.id
		WHERE g.is_private_chat = 1
		GROUP BY
			g.id,
			u.id,
			u.first_name,
			u.last_name,
			p.avatar_path
		ORDER BY
			MAX(m.created_at) DESC,
			g.created_at DESC
		LIMIT 20 OFFSET ?
	`, userID, userID, offset)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var chat models.PrivateChat

		err := rows.Scan(
			&chat.UserID,
			&chat.FirstName,
			&chat.LastName,
			&chat.Avatar,
			&chat.GroupID,
		)

		if err != nil {
			return nil, err
		}

		chats = append(chats, chat)
	}

	return chats, rows.Err()
}

func SearchChatUsers(db *sql.DB, userID, offset int, search string) ([]models.PrivateChat, error) {
	chats := make([]models.PrivateChat, 0)

	pattern := "%" + search + "%"

	rows, err := db.Query(`
		SELECT
			u.id,
			u.first_name,
			u.last_name,
			p.avatar_path,
			COALESCE(MAX(x.group_id), 0)
		FROM (
			SELECT gu2.user_id AS user_id, g.id AS group_id
			FROM groups g
			JOIN groups_users gu1
				ON gu1.group_id = g.id
				AND gu1.user_id = ?
			JOIN groups_users gu2
				ON gu2.group_id = g.id
				AND gu2.user_id != ?
			WHERE g.is_private_chat = 1

			UNION

			SELECT follower_id AS user_id, NULL AS group_id
			FROM user_followers
			WHERE target_id = ?
				AND status = 1

			UNION

			SELECT target_id AS user_id, NULL AS group_id
			FROM user_followers
			WHERE follower_id = ?
				AND status = 1
		) x
		JOIN user u ON u.id = x.user_id
		JOIN profile p ON p.user_id = u.id
		WHERE u.id != ?
			AND (
				u.first_name LIKE ?
				OR u.last_name LIKE ?
				OR (u.first_name || ' ' || u.last_name) LIKE ?
			)
		GROUP BY
			u.id,
			u.first_name,
			u.last_name,
			p.avatar_path
		ORDER BY
			CASE WHEN MAX(x.group_id) IS NULL THEN 1 ELSE 0 END,
			u.first_name,
			u.last_name
		LIMIT 15 OFFSET ?
	`,
		userID,
		userID,
		userID,
		userID,
		userID,
		pattern,
		pattern,
		pattern,
		offset,
	)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var chat models.PrivateChat

		err := rows.Scan(
			&chat.UserID,
			&chat.FirstName,
			&chat.LastName,
			&chat.Avatar,
			&chat.GroupID,
		)

		if err != nil {
			return nil, err
		}

		chats = append(chats, chat)
	}

	return chats, rows.Err()
}

func PrivateChatExists(db *sql.DB, groupID int) (bool, error) {
	var exists int

	err := db.QueryRow(`
		SELECT 1
		FROM groups
		WHERE id = ?
			AND is_private_chat = 1
	`, groupID).Scan(&exists)

	if err == sql.ErrNoRows {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

func MakePrivateChat(db *sql.DB, userID, targetID int) (int, error) {
	res, err := db.Exec(`
	INSERT INTO groups (is_private_chat)
	VALUES (1)`)
	if err != nil {
		return -1, err
	}

	groupID, err := res.LastInsertId()
	if err != nil {
		return -1, err
	}

	_, err = db.Exec(`
	INSERT INTO groups_users (group_id, user_id)
	VALUES (?, ?)
`, groupID, userID)

	if err != nil {
		return -1, err
	}

	_, err = db.Exec(`
	INSERT INTO groups_users (group_id, user_id)
	VALUES (?, ?)
`, groupID, targetID)

	if err != nil {
		return -1, err
	}

	return int(groupID), nil
}

func UserInGroup(db *sql.DB, userID, groupID int) (bool, error) {
	var exists int

	err := db.QueryRow(`
        SELECT 1
        FROM groups_users
        WHERE user_id = ? AND group_id = ? AND COALESCE(status, 1) = 1
        LIMIT 1
    `, userID, groupID).Scan(&exists)

	if err == sql.ErrNoRows {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

func GetGroupMembersIds(db *sql.DB, groupID int) ([]int, error) {
	var ids []int
	rows, err := db.Query(`
		SELECT user_id FROM groups_users WHERE group_id = ? AND COALESCE(status, 1) = 1
	`, groupID)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	return ids, nil
}

// GetGroupName returns the name of a group chat ("" when it has none).
func GetGroupName(db *sql.DB, groupID int) (string, error) {
	var name string

	err := db.QueryRow(`
		SELECT COALESCE(name, '') FROM groups WHERE id = ?
	`, groupID).Scan(&name)

	return name, err
}
