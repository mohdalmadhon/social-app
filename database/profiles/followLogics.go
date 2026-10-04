package profiles

import (
	"database/sql"
	"fmt"
	"social/internal/models"
)

func SendFollowRequest(db *sql.DB, targetID, followerID, requestCode int) error {
	if targetID == followerID {
		return fmt.Errorf("user cannot follow themselves")
	}

	if requestCode != -1 && requestCode != 0 && requestCode != 1 {
		return fmt.Errorf("invalid follow request code: %d", requestCode)
	}

	if requestCode == -1 {
		_, err := db.Exec(`
			DELETE FROM user_followers
			WHERE target_id = ? AND follower_id = ?
		`, targetID, followerID)

		return err
	}

	var status sql.NullInt64

	err := db.QueryRow(`
		SELECT status
		FROM user_followers
		WHERE target_id = ? AND follower_id = ?
	`, targetID, followerID).Scan(&status)

	if err == sql.ErrNoRows {
		_, err = db.Exec(`
			INSERT INTO user_followers (
				target_id,
				follower_id,
				status
			)
			VALUES (?, ?, ?)
		`, targetID, followerID, requestCode)

		return err
	}

	if err != nil {
		return err
	}

	_, err = db.Exec(`
		UPDATE user_followers
		SET status = ?
		WHERE target_id = ? AND follower_id = ?
	`, requestCode, targetID, followerID)

	return err
}

func GetFollowers(db *sql.DB, targetID, count, offset int) (map[int]models.UserRegistration, error) {
	rows, err := db.Query(`
		SELECT
			u.id,
			u.first_name,
			u.last_name,
			p.avatar_path
		FROM user_followers uf
		JOIN user u
			ON u.id = uf.follower_id
		LEFT JOIN profile p
			ON p.user_id = u.id
		WHERE uf.target_id = ?
		ORDER BY u.id
		LIMIT ?
		OFFSET ?
	`, targetID, count, offset)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	followers := make(map[int]models.UserRegistration)

	for rows.Next() {
		var id int
		var follower models.UserRegistration
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
			return nil, err
		}

		if firstName.Valid {
			follower.FirstName = firstName.String
		} else {
			follower.FirstName = ""
		}

		if lastName.Valid {
			follower.LastName = lastName.String
		} else {
			follower.LastName = ""
		}

		if avatar.Valid {
			follower.Avatar = avatar.String
		} else {
			follower.Avatar = ""
		}

		followers[id] = follower
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return followers, nil
}

func GetFollowing(db *sql.DB, followerID, count, offset int) (map[int]models.UserRegistration, error) {
	rows, err := db.Query(`
		SELECT
			u.id,
			u.first_name,
			u.last_name,
			p.avatar_path
		FROM user_followers uf
		JOIN user u
			ON u.id = uf.target_id
		LEFT JOIN profile p
			ON p.user_id = u.id
		WHERE uf.follower_id = ?
		LIMIT ?
	`, followerID, count)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	following := make(map[int]models.UserRegistration)

	for rows.Next() {
		var id int
		var user models.UserRegistration
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
			return nil, err
		}

		if firstName.Valid {
			user.FirstName = firstName.String
		} else {
			user.FirstName = ""
		}

		if lastName.Valid {
			user.LastName = lastName.String
		} else {
			user.LastName = ""
		}

		if avatar.Valid {
			user.Avatar = avatar.String
		} else {
			user.Avatar = ""
		}

		following[id] = user
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return following, nil
}

func SearchFollows(db *sql.DB, userID int, searchValue string) ([]models.UserRegistration, error) {
	searchPattern := "%" + searchValue + "%"

	rows, err := db.Query(`
        SELECT u.id, u.first_name, u.last_name, p.avatar_path
        FROM user u
        LEFT JOIN profile p
            ON p.user_id = u.id
        WHERE EXISTS (
            SELECT 1
            FROM user_followers uf
            WHERE uf.target_id = u.id
            AND uf.follower_id = ?
        )
        AND (
            u.first_name LIKE ?
            OR u.last_name LIKE ?
        )
        LIMIT 100
    `, userID, searchPattern, searchPattern)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	follows := make([]models.UserRegistration, 0)

	for rows.Next() {
		var user models.UserRegistration

		err := rows.Scan(
			&user.ID,
			&user.FirstName,
			&user.LastName,
			&user.Avatar,
		)

		if err != nil {
			return nil, err
		}

		follows = append(follows, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return follows, nil
}

func SearchFollowing(db *sql.DB, userID int, searchValue string) ([]models.UserRegistration, error) {
	searchPattern := "%" + searchValue + "%"

	rows, err := db.Query(`
        SELECT u.id, u.first_name, u.last_name, p.avatar_path
        FROM user u
        LEFT JOIN profile p
            ON p.user_id = u.id
        WHERE EXISTS (
            SELECT 1
            FROM user_followers uf
            WHERE uf.follower_id = u.id
            AND uf.target_id = ?
        )
        AND (
            u.first_name LIKE ?
            OR u.last_name LIKE ?
        )
        LIMIT 100
    `, userID, searchPattern, searchPattern)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	following := make([]models.UserRegistration, 0)

	for rows.Next() {
		var user models.UserRegistration

		err := rows.Scan(
			&user.ID,
			&user.FirstName,
			&user.LastName,
			&user.Avatar,
		)

		if err != nil {
			return nil, err
		}

		following = append(following, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return following, nil
}

func AcceptFollowRequest(db *sql.DB, requesterID int, targetID int) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var exists int

	err = tx.QueryRow(`
		SELECT 1
		FROM user_followers
		WHERE follower_id = ?
		AND target_id = ?
		AND status = 0
	`, requesterID, targetID).Scan(&exists)

	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		UPDATE user_followers
		SET status = 1
		WHERE follower_id = ?
		AND target_id = ?
		AND status = 0
	`, requesterID, targetID)

	if err != nil {
		return err
	}

	var notificationID int

	err = tx.QueryRow(`
		SELECT nt.notifications_id
		FROM notifications_types nt
		JOIN notifications n
			ON n.id = nt.notifications_id
		WHERE nt.follow_request_user_id = ?
		AND n.user_id = ?
	`, requesterID, targetID).Scan(&notificationID)
	
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		DELETE FROM notifications
		WHERE id = ?
	`, notificationID)

	if err != nil {
		return err
	}

	return tx.Commit()
}

func RejectFollowRequest(db *sql.DB, requesterID int, targetID int) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	result, err := tx.Exec(`
		DELETE FROM user_followers
		WHERE follower_id = ?
		AND target_id = ?
		AND status = 0
	`, requesterID, targetID)

	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	_, err = tx.Exec(`
		DELETE FROM notifications
		WHERE user_id = ?
		AND id IN (
			SELECT notifications_id
			FROM notifications_types
			WHERE follow_request_user_id = ?
		)
	`, targetID, requesterID)

	if err != nil {
		return err
	}

	return tx.Commit()
}
