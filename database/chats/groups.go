package chats

import (
	"database/sql"
	"log"
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

		canMessage, err := CanSendMessage(db, userID, chat.UserID)
		if err != nil {
			return nil, err
		}

		if !canMessage {
			hasPrivateChat, err := HasPrivateChat(db, userID, chat.UserID)
			if err != nil {
				return nil, err
			}
			
			if hasPrivateChat != 1 {
				continue
			} 
		}
		
		chats = append(chats, chat)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	log.Println(len(chats))
	return chats, nil
}

func SearchChatUsers(db *sql.DB, userID, offset int, search string) ([]models.PrivateChat, error) {
	var chats []models.PrivateChat

	rows, err := db.Query(`
		SELECT
			u.id,
			u.first_name,
			u.last_name,
			p.avatar_path,
			COALESCE(x.group_id, 0)
		FROM (
			SELECT gu2.user_id, g.id AS group_id, 1 AS priority
			FROM groups g
			JOIN groups_users gu1 ON g.id = gu1.group_id
			JOIN groups_users gu2 ON g.id = gu2.group_id
			WHERE g.is_private_chat = 1
				AND gu1.user_id = ?
				AND gu2.user_id != ?

			UNION

			SELECT uf1.target_id, NULL, 2
			FROM user_followers uf1
			JOIN user_followers uf2
				ON uf1.target_id = uf2.follower_id
				AND uf1.follower_id = uf2.target_id
			WHERE uf1.follower_id = ?
				AND uf1.status = 1
				AND uf2.status = 1

			UNION

			SELECT target_id, NULL, 3
			FROM user_followers
			WHERE follower_id = ?
				AND status = 1
		) x
		JOIN user u ON u.id = x.user_id
		JOIN profile p ON p.user_id = u.id
		WHERE u.first_name LIKE ?
			OR u.last_name LIKE ?
		ORDER BY x.priority, u.first_name, u.last_name
		LIMIT 15 OFFSET ?
	`,
		userID,
		userID,
		userID,
		userID,
		"%"+search+"%",
		"%"+search+"%",
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

		canMessage, err := CanSendMessage(db, userID, chat.UserID)
		if err != nil {
			return nil, err
		}

		if !canMessage {
			hasChat, err := HasPrivateChat(db, userID, chat.UserID)
			if err != nil {
				return nil, err
			}

			if hasChat != 1 {
				continue
			}
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
        WHERE user_id = ? AND group_id = ?
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
		SELECT user_id FROM groups_users WHERE group_id = ?
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
