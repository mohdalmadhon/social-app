package chats

import (
	"database/sql"
	"social/database/users"
	"social/internal/models"
)

func GetMessages(db *sql.DB, userID, groupID, offset int) ([]models.Message, error) {
	var messages []models.Message
	var exists int

	err := db.QueryRow(`
		SELECT 1
		FROM groups_users
		WHERE group_id = ?
			AND user_id = ?
	`, groupID, userID).Scan(&exists)

	if err != nil {
		return nil, err
	}

	rows, err := db.Query(`
		SELECT
			u.id,
			u.first_name,
			u.last_name,
			p.avatar_path,
			m.content,
			m.created_at,
			m.id
		FROM messages m
		JOIN user u ON u.id = m.sender_id
		JOIN profile p ON p.user_id = u.id
		WHERE m.group_id = ?
		ORDER BY m.created_at DESC, m.id DESC
		LIMIT 20 OFFSET ?
	`, groupID, offset)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var m models.Message

		err := rows.Scan(
			&m.Sender.ID,
			&m.Sender.FirstName,
			&m.Sender.LastName,
			&m.Sender.Avatar,
			&m.Content,
			&m.CreatedAt,
			&m.ID,
		)

		if err != nil {
			return nil, err
		}

		if m.Sender.ID == userID {
			m.Sender.ID = -1
		}

		messages = append(messages, m)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

func ChatExists(db *sql.DB, groupID int) (bool, error) {
	var exists int

	err := db.QueryRow(`
		SELECT 1
		FROM groups
		WHERE id = ?
	`, groupID).Scan(&exists)

	if err == sql.ErrNoRows {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

func HasPrivateChat(db *sql.DB, user1, user2 int) (int, error) {
	var groupID int

	err := db.QueryRow(`
		SELECT g.id
		FROM groups g
		WHERE g.is_private_chat = 1
		AND EXISTS (
			SELECT 1
			FROM groups_users gu1
			WHERE gu1.group_id = g.id
			AND gu1.user_id = ?
		)
		AND EXISTS (
			SELECT 1
			FROM groups_users gu2
			WHERE gu2.group_id = g.id
			AND gu2.user_id = ?
		)
		LIMIT 1
	`, user1, user2).Scan(&groupID)

	if err == sql.ErrNoRows {
		return -1, nil
	}

	if err != nil {
		return -1, err
	}

	return groupID, nil
}

func AddMessages(db *sql.DB, content string, userID, groupID int) error {
	_, err := db.Exec(`
		INSERT INTO MESSAGES (content, sender_id, group_id)
		VALUES (?,?,?)
	`, content, userID, groupID)
	return err
}

func CanSendMessage(db *sql.DB, userID, targetID int) (bool, error) {
	if userID <= 0 || targetID <= 0 || userID == targetID {
		return false, nil
	}

	var preference string

	err := db.QueryRow(`
		SELECT chat FROM user_preferences WHERE user_id = ?
	`, targetID).Scan(&preference)

	if err == sql.ErrNoRows {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	switch preference {
	case "any":
		return true, nil

	case "none":
		return false, nil

	case "following":
		return users.IsFollowing(db, targetID, userID)

	case "friends":
		return users.IsFriend(db, userID, targetID)

	case "following-followers":
		targetFollowsUser, err := users.IsFollowing(db, targetID, userID)
		if err != nil {
			return false, err
		}

		if targetFollowsUser {
			return true, nil
		}

		return users.IsFollowing(db, userID, targetID)

	case "friends-following":
		friend, err := users.IsFriend(db, userID, targetID)
		if err != nil {
			return false, err
		}

		if friend {
			return true, nil
		}

		return users.IsFollowing(db, targetID, userID)
	}

	return false, nil
}

func IsPrivateChat(db *sql.DB, groupID, userID int) (bool, int, error) {
	var isPrivate int

	err := db.QueryRow(`
		SELECT is_private_chat
		FROM groups
		WHERE id = ?
	`, groupID).Scan(&isPrivate)

	if err != nil {
		return false, -1, err
	}

	if isPrivate != 1 {
		return false, -1, nil
	}

	var targetID int

	err = db.QueryRow(`
		SELECT user_id
		FROM groups_users
		WHERE group_id = ? AND user_id <> ?
		LIMIT 1
	`, groupID, userID).Scan(&targetID)

	if err != nil {
		return true, -1, err
	}

	return true, targetID, nil
}

func GetChatMeta(db *sql.DB, groupID int) (bool, string, error) {
	var isPrivate int
	var name sql.NullString

	err := db.QueryRow(`
		SELECT is_private_chat, name
		FROM groups
		WHERE id = ?
	`, groupID).Scan(&isPrivate, &name)

	if err != nil {
		return false, "", err
	}

	return isPrivate == 1, name.String, nil
}
