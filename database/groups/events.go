package groups

import (
	"database/sql"
	"social/internal/models"
)

func InsertGroupEvent(db *sql.DB, event models.NewGroupEvent, userID int) (int, error) {
	result, err := db.Exec(`
		INSERT INTO group_events (group_id, user_id, title, description, event_time)
		SELECT ?, ?, ?, ?, ?
		WHERE EXISTS (
			SELECT 1
			FROM groups_users
			WHERE group_id = ?
			AND user_id = ?
			AND status = 1
		)
	`,
		event.GroupID,
		userID,
		event.Title,
		event.Description,
		event.EventTime,
		event.GroupID,
		userID,
	)

	if err != nil {
		return 0, err
	}

	rows, err := result.RowsAffected()

	if err != nil {
		return 0, err
	}

	if rows == 0 {
		return 0, sql.ErrNoRows
	}

	id, err := result.LastInsertId()

	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func GetGroupEvent(db *sql.DB, eventID, userID int) (models.GroupEvent, error) {
	var event models.GroupEvent
	var response sql.NullInt64

	err := db.QueryRow(`
		SELECT
			e.id,
			e.group_id,
			e.title,
			e.description,
			e.event_time,
			e.created_at,
			u.id,
			u.first_name,
			u.last_name,
			COALESCE(p.avatar_path, ''),
			(
				SELECT COUNT(*)
				FROM group_event_responses r
				WHERE r.event_id = e.id
				AND r.response = 1
			),
			(
				SELECT COUNT(*)
				FROM group_event_responses r
				WHERE r.event_id = e.id
				AND r.response = 0
			),
			(
				SELECT r2.response
				FROM group_event_responses r2
				WHERE r2.event_id = e.id
				AND r2.user_id = ?
			)
		FROM group_events e
		JOIN user u ON u.id = e.user_id
		LEFT JOIN profile p ON p.user_id = u.id
		WHERE e.id = ?
	`, userID, eventID).Scan(
		&event.ID,
		&event.GroupID,
		&event.Title,
		&event.Description,
		&event.EventTime,
		&event.CreatedAt,
		&event.Creator.ID,
		&event.Creator.FirstName,
		&event.Creator.LastName,
		&event.Creator.Avatar,
		&event.GoingCount,
		&event.NotGoingCount,
		&response,
	)

	if err != nil {
		return event, err
	}

	if response.Valid {
		value := int(response.Int64)
		event.UserResponse = &value
	}

	return event, nil
}

func GetGroupEvents(db *sql.DB, groupID, userID, limit, offset int) ([]models.GroupEvent, error) {
	rows, err := db.Query(`
		SELECT
			e.id,
			e.group_id,
			e.title,
			e.description,
			e.event_time,
			e.created_at,
			u.id,
			u.first_name,
			u.last_name,
			COALESCE(p.avatar_path, ''),
			(
				SELECT COUNT(*)
				FROM group_event_responses r
				WHERE r.event_id = e.id
				AND r.response = 1
			),
			(
				SELECT COUNT(*)
				FROM group_event_responses r
				WHERE r.event_id = e.id
				AND r.response = 0
			),
			(
				SELECT r2.response
				FROM group_event_responses r2
				WHERE r2.event_id = e.id
				AND r2.user_id = ?
			)
		FROM group_events e
		JOIN user u ON u.id = e.user_id
		LEFT JOIN profile p ON p.user_id = u.id
		WHERE e.group_id = ?
		ORDER BY e.event_time DESC, e.id DESC
		LIMIT ?
		OFFSET ?
	`, userID, groupID, limit, offset)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	events := []models.GroupEvent{}

	for rows.Next() {
		var event models.GroupEvent
		var response sql.NullInt64

		err := rows.Scan(
			&event.ID,
			&event.GroupID,
			&event.Title,
			&event.Description,
			&event.EventTime,
			&event.CreatedAt,
			&event.Creator.ID,
			&event.Creator.FirstName,
			&event.Creator.LastName,
			&event.Creator.Avatar,
			&event.GoingCount,
			&event.NotGoingCount,
			&response,
		)

		if err != nil {
			return nil, err
		}

		if response.Valid {
			value := int(response.Int64)
			event.UserResponse = &value
		}

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

func EventGroupID(db *sql.DB, eventID int) (int, error) {
	var groupID int

	err := db.QueryRow(`
		SELECT group_id
		FROM group_events
		WHERE id = ?
	`, eventID).Scan(&groupID)

	if err != nil {
		return 0, err
	}

	return groupID, nil
}

func SetGroupEventResponse(db *sql.DB, eventID, userID, response int) error {
	var current sql.NullInt64

	err := db.QueryRow(`
		SELECT response
		FROM group_event_responses
		WHERE event_id = ?
		AND user_id = ?
	`, eventID, userID).Scan(&current)

	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if err == sql.ErrNoRows {
		_, err = db.Exec(`
			INSERT INTO group_event_responses (event_id, user_id, response)
			VALUES (?, ?, ?)
		`, eventID, userID, response)

		return err
	}

	if current.Valid && int(current.Int64) == response {
		_, err = db.Exec(`
			DELETE FROM group_event_responses
			WHERE event_id = ?
			AND user_id = ?
		`, eventID, userID)

		return err
	}

	_, err = db.Exec(`
		UPDATE group_event_responses
		SET response = ?
		WHERE event_id = ?
		AND user_id = ?
	`, response, eventID, userID)

	return err
}

func GetGroupEventVoters(db *sql.DB, eventID, response, limit, offset int) ([]models.GroupEventVoter, error) {
	rows, err := db.Query(`
		SELECT
			u.id,
			u.first_name,
			u.last_name,
			COALESCE(u.username, ''),
			COALESCE(p.avatar_path, ''),
			r.response,
			r.created_at
		FROM group_event_responses r
		JOIN user u ON u.id = r.user_id
		LEFT JOIN profile p ON p.user_id = u.id
		WHERE r.event_id = ?
		AND r.response = ?
		ORDER BY r.created_at DESC, u.id DESC
		LIMIT ?
		OFFSET ?
	`, eventID, response, limit, offset)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	voters := []models.GroupEventVoter{}

	for rows.Next() {
		var voter models.GroupEventVoter

		err := rows.Scan(
			&voter.User.ID,
			&voter.User.FirstName,
			&voter.User.LastName,
			&voter.User.UserName,
			&voter.User.Avatar,
			&voter.Response,
			&voter.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		voters = append(voters, voter)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return voters, nil
}

// GetActiveMemberIDs returns the ids of the users who are accepted members of
// a group (pending / declined invites are left out).
func GetActiveMemberIDs(db *sql.DB, groupID int) ([]int, error) {
	rows, err := db.Query(`
		SELECT user_id
		FROM groups_users
		WHERE group_id = ?
		AND status = 1
	`, groupID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var ids []int

	for rows.Next() {
		var id int

		if err := rows.Scan(&id); err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	return ids, rows.Err()
}
