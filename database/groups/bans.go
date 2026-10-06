package groups

import "database/sql"

func IsBanned(db *sql.DB, groupID, userID int) (bool, error) {
	var exists bool

	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1
			FROM group_bans
			WHERE group_id = ?
				AND user_id = ?
		)
	`, groupID, userID).Scan(&exists)

	return exists, err
}

func KickMember(db *sql.DB, groupID, userID, kickedBy int) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	res, err := tx.Exec(`
		DELETE FROM groups_users
		WHERE group_id = ?
			AND user_id = ?
			AND status = 1
	`, groupID, userID)

	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return sql.ErrNoRows
	}

	_, err = tx.Exec(`
		INSERT OR IGNORE INTO group_bans (group_id, user_id, banned_by)
		VALUES (?, ?, ?)
	`, groupID, userID, kickedBy)

	if err != nil {
		return err
	}

	return tx.Commit()
}

func LeaveGroup(db *sql.DB, groupID, userID int) error {
	res, err := db.Exec(`
		DELETE FROM groups_users
		WHERE group_id = ?
			AND user_id = ?
			AND status = 1
	`, groupID, userID)

	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func LiftBan(db *sql.DB, groupID, userID int) error {
	_, err := db.Exec(`
		DELETE FROM group_bans
		WHERE group_id = ?
			AND user_id = ?
	`, groupID, userID)

	return err
}
