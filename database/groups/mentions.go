package groups

import "database/sql"

type Mentionable struct {
	ID        int    `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	UserName  string `json:"username"`
	Avatar    string `json:"avatar"`
}

func SearchMentionable(db *sql.DB, groupID, excludeUserID int, search string, limit int) ([]Mentionable, error) {
	pattern := "%" + search + "%"

	rows, err := db.Query(`
		SELECT
			u.id,
			COALESCE(u.first_name, ''),
			COALESCE(u.last_name, ''),
			u.username,
			COALESCE(p.avatar_path, '')
		FROM groups_users gu
		JOIN user u ON u.id = gu.user_id
		LEFT JOIN profile p ON p.user_id = u.id
		WHERE gu.group_id = ?
			AND gu.status = 1
			AND u.id != ?
			AND u.username IS NOT NULL
			AND u.username != ''
			AND (
				? = ''
				OR u.username LIKE ?
				OR u.first_name LIKE ?
				OR u.last_name LIKE ?
			)
		ORDER BY
			CASE WHEN u.username LIKE ? THEN 0 ELSE 1 END,
			u.first_name,
			u.last_name
		LIMIT ?
	`, groupID, excludeUserID, search, pattern, pattern, pattern, search+"%", limit)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	result := []Mentionable{}

	for rows.Next() {
		var m Mentionable

		if err := rows.Scan(&m.ID, &m.FirstName, &m.LastName, &m.UserName, &m.Avatar); err != nil {
			return nil, err
		}

		result = append(result, m)
	}

	return result, rows.Err()
}
