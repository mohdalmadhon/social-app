package groups

import (
	"database/sql"
	"encoding/json"
	"log"
	"social/database/chats"
	"social/database/users"
	"social/internal/models"
)

func SendInvites(db *sql.DB, targetID int, g models.Group) error {
	u, err := users.GetUserSimpleData(db, g.UserID)
	if err != nil {
		return err
	}
	
	_, err = db.Exec(`INSERT INTO groups_users (group_id, user_id, status) VALUES (?,?,0)`, g.ID, targetID)
	if err != nil {
		return err
	}

	payload := map[string]any{
		"type": "invite",
		"group": map[string]any{
			"id":     g.ID,
			"name":   g.Title,
			"avatar": g.Avatar,
		},
		"user": map[string]any{
			"id":        u.ID,
			"firstName": u.FirstName,
			"lastName":  u.LastName,
			"avatar":    u.Avatar,
		},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	gID, err := chats.HasPrivateChat(db, g.UserID, targetID)
	if err != nil {
		return err
	}

	if gID == -1 {
		gID, err = chats.MakePrivateChat(db, g.UserID, targetID)
		if err != nil {
			return err
		}
	}

	return chats.AddMessages(db, string(data), g.UserID, gID)
}

func ChangeStatus(db *sql.DB, userID, status, groupID int) error {
	log.Println(status)
	if status == -1 {
		_, err := db.Exec(`
			DELETE FROM groups_users
			WHERE group_id = ?
			AND user_id = ?
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