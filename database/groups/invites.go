package groups

import (
	"database/sql"
	"log"
)

func AddPendingMember(db *sql.DB, groupID, userID, inviterID int) error {
	_, err := db.Exec(`
		INSERT INTO groups_users (group_id, user_id, status, invited_by)
		VALUES (?, ?, 0, ?)
	`, groupID, userID, inviterID)

	return err
}

func RemoveMember(db *sql.DB, groupID, userID int) error {
	_, err := db.Exec(`
		DELETE FROM groups_users
		WHERE group_id = ?
			AND user_id = ?
	`, groupID, userID)

	return err
}

func ChangeStatus(db *sql.DB, userID, status, groupID int) error {
	if status == -1 {
		_, err := db.Exec(`
			DELETE FROM groups_users
			WHERE group_id = ?
			AND user_id = ?
			AND status = 0
			AND invited_by IS NOT NULL
		`, groupID, userID)

		return err
	}

	if status != 1 {
		return nil
	}

	res, err := db.Exec(`
		UPDATE groups_users
		SET status = ?
		WHERE user_id = ?
		AND group_id = ?
		AND status = 0
		AND invited_by IS NOT NULL
	`, status, userID, groupID)

	log.Println(res)
	return err
}

func DeleteInvite(db *sql.DB, userID, senderID, groupID int) error {
	_, err := db.Exec(`
		DELETE FROM messages
		WHERE id = (
			SELECT m.id
			FROM messages m
			JOIN groups g
				ON g.id = m.group_id
			JOIN groups_users gu1
				ON gu1.group_id = g.id
			JOIN groups_users gu2
				ON gu2.group_id = g.id
			WHERE g.is_private_chat = 1
			AND gu1.user_id = ?
			AND gu2.user_id = ?
			AND m.sender_id = ?
			AND json_extract(m.content, '$.type') = 'invite'
			AND json_extract(m.content, '$.group.id') = ?
			ORDER BY m.id DESC
			LIMIT 1
		)
	`, userID, senderID, senderID, groupID)

	return err
}
