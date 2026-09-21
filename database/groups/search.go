package groups

import (
	"database/sql"

	"social/database/dbutil"
	"social/internal/models"
)

// SearchGroups finds group chats whose name or description contains the
// search text. Groups the searching user belongs to come first.
func SearchGroups(db *sql.DB, userID int, search string, limit, offset int) ([]models.SearchGroup, error) {
	pattern := dbutil.LikePattern(search)

	rows, err := db.Query(`
		SELECT
			g.id,
			g.name,
			g.description,
			g.avatar,
			(
				SELECT COUNT(*)
				FROM groups_users gu2
				WHERE gu2.group_id = g.id
			) AS members_count,
			EXISTS (
				SELECT 1
				FROM groups_users gu
				WHERE gu.group_id = g.id
					AND gu.user_id = ?
			) AS is_member
		FROM groups g
		WHERE g.is_private_chat = 0
			AND (
				g.name LIKE ? ESCAPE '\'
				OR g.description LIKE ? ESCAPE '\'
			)
		ORDER BY
			is_member DESC,
			g.name COLLATE NOCASE,
			g.id
		LIMIT ? OFFSET ?
	`, userID, pattern, pattern, limit, offset)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	result := make([]models.SearchGroup, 0)

	for rows.Next() {
		var g models.SearchGroup
		var name sql.NullString
		var description sql.NullString
		var avatar sql.NullString

		if err := rows.Scan(
			&g.ID,
			&name,
			&description,
			&avatar,
			&g.MembersCount,
			&g.IsMember,
		); err != nil {
			return nil, err
		}

		g.Name = name.String
		g.Description = description.String
		g.Avatar = avatar.String

		result = append(result, g)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
