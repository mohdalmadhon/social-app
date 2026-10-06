package groups

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"social/internal/helpers"
	"social/internal/models"
	"strconv"
	"strings"
)

func GetGroups(db *sql.DB, userID int) ([]models.Group, error) {
	var groups []models.Group

	rows, err := db.Query(`
		SELECT id,name, users
		FROM user_posts_groups
		WHERE user_id = ?
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var users string
		var group models.Group

		err := rows.Scan(
			&group.ID,
			&group.Name,
			&users,
		)

		if err != nil {
			return nil, err
		}
		if group.ID == -1 || group.ID == 0 {
			continue
		}
		log.Println(users)
		if users != "" {
			usersIDsStr := strings.Split(users, ":")
			usersIDs := make([]int, 0, len(usersIDsStr))

			for _, idStr := range usersIDsStr {
				id, err := strconv.Atoi(idStr)
				if err != nil {
					return nil, errors.New("failed to get users")
				}

				usersIDs = append(usersIDs, id)
			}

			group.Users, err = helpers.GetPostGroupUsers(db, usersIDs)
			if err != nil {
				return nil, err
			}
		}

		groups = append(groups, group)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return groups, nil
}

func AddGroup(db *sql.DB, group models.NewGroup) error {
	users := helpers.JoinUserIDs(group.Users)
	_, err := db.Exec(`insert into user_posts_groups (name, user_id, users)
					   VALUES (?,?,?)`, group.Name, group.UserID, users)
	return err
}

func DeleteGroup(db *sql.DB, groupID, userID int) error {
	_, err := db.Exec(`DELETE FROM user_posts_groups 
					   WHERE user_id = ? AND id = ?`, userID, groupID)
	return err
}

func UpdateGroup(db *sql.DB, group models.NewGroup) error {
	var current string
	err := db.QueryRow(`select users from user_posts_groups where user_id = ? AND name = ?`,
		group.UserID, group.Name).Scan(&current)
	if err != nil {
		return err
	}

	users := helpers.JoinUserIDs(group.Users)
	if users == current {
		return nil
	}

	_, err = db.Exec(`
		UPDATE user_posts_groups
		SET users = ?
		WHERE user_id = ? AND name = ?
	`, users, group.UserID, group.Name)
	if err != nil {
		return err
	}
	return nil
}

func MakeNewGroup(db *sql.DB, g models.Group, userIDs []int) (models.Group, []int, error) {
	result, err := db.Exec(`
		INSERT INTO groups (
			is_private_chat,
			owner_id,
			name,
			description,
			avatar
		)
		VALUES (?, ?, ?, ?, ?)
	`,
		0,
		g.UserID,
		g.Title,
		g.Description,
		g.Avatar,
	)
	if err != nil {
		return g, nil, err
	}

	groupID, err := result.LastInsertId()
	if err != nil {
		return g, nil, err
	}

	g.ID = int(groupID)

	_, err = db.Exec(`
		INSERT INTO groups_users (group_id, user_id, status)
		VALUES (?, ?, 1)
	`, groupID, g.UserID)
	if err != nil {
		return g, nil, err
	}

	return g, userIDs, nil
}

func AddMembers(db *sql.DB, groupID, userID int) error {
	_, err := db.Exec(`
			INSERT INTO groups_users (group_id, user_id, status)
			VALUES (?,?, 1)		
		`, groupID, userID)
	return err
}

func SearchInvites(db *sql.DB, userID, groupID int, searchValue string) ([]models.UserRegistration, error) {
	search := "%" + searchValue + "%"

	users := make([]models.UserRegistration, 0)

	query := `
		SELECT
			u.id,
			u.first_name,
			u.last_name,
			p.avatar_path
		FROM user u
		JOIN profile p ON p.user_id = u.id
		WHERE u.id != ?
			AND (
				u.username LIKE ?
				OR u.first_name LIKE ?
				OR u.last_name LIKE ?
			)
	`

	args := []any{
		userID,
		search,
		search,
		search,
	}

	query += `
		AND u.id IN (
			SELECT target_id
			FROM user_followers
			WHERE follower_id = ?
				AND status = 1

			UNION

			SELECT follower_id
			FROM user_followers
			WHERE target_id = ?
				AND status = 1
		)
	`

	args = append(args, userID, userID)

	query += `
		AND CASE COALESCE(
			(SELECT up.group_invite FROM user_preferences up WHERE up.user_id = u.id),
			'following'
		)
			WHEN 'none' THEN 0
			WHEN 'friends' THEN (
				EXISTS (
					SELECT 1 FROM user_followers a
					WHERE a.follower_id = ? AND a.target_id = u.id AND a.status = 1
				)
				AND EXISTS (
					SELECT 1 FROM user_followers b
					WHERE b.follower_id = u.id AND b.target_id = ? AND b.status = 1
				)
			)
			WHEN 'following' THEN 1
			ELSE 0
		END
	`

	args = append(args, userID, userID)

	if groupID != -1 {
		query += `
			AND NOT EXISTS (
				SELECT 1
				FROM groups_users gu
				WHERE gu.group_id = ?
					AND gu.user_id = u.id
			)
		`

		args = append(args, groupID)
	}

	query += `
		ORDER BY
			CASE
				WHEN EXISTS (
					SELECT 1
					FROM user_followers uf1
					WHERE uf1.follower_id = ?
						AND uf1.target_id = u.id
				)
				AND EXISTS (
					SELECT 1
					FROM user_followers uf2
					WHERE uf2.follower_id = u.id
						AND uf2.target_id = ?
				)
				THEN 0
				ELSE 1
			END,
			u.first_name,
			u.last_name
		LIMIT 20
	`

	args = append(args, userID, userID)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var u models.UserRegistration

		if err := rows.Scan(
			&u.ID,
			&u.FirstName,
			&u.LastName,
			&u.Avatar,
		); err != nil {
			return nil, err
		}

		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func GetGroupChats(db *sql.DB, userID, offset int) ([]models.Group, error) {
	groups := make([]models.Group, 0)

	rows, err := db.Query(`
		SELECT 
			g.id,
			g.name,
			g.avatar,
			(
				SELECT COUNT(*)
				FROM groups_users gu2
				WHERE gu2.group_id = g.id
			)
		FROM groups g
		WHERE EXISTS (
			SELECT 1
			FROM groups_users gu
			WHERE gu.group_id = g.id
			  AND gu.user_id = ?
			  AND status = 1
		)
		AND g.is_private_chat = 0
		ORDER BY g.name, g.id
		LIMIT 12 OFFSET ?
	`, userID, offset)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var g models.Group

		if err := rows.Scan(
			&g.ID,
			&g.Title,
			&g.Avatar,
			&g.Count,
		); err != nil {
			return nil, err
		}

		groups = append(groups, g)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return groups, nil
}

func DiscoverGroups(db *sql.DB, userID, offset int, search string) ([]models.Group, error) {
	groups := make([]models.Group, 0)

	query := `
		SELECT
			g.id,
			g.name,
			g.description,
			g.avatar,
			(
				SELECT COUNT(*)
				FROM groups_users gu2
				WHERE gu2.group_id = g.id
			) AS members_count
		FROM groups g
		WHERE NOT EXISTS (
			SELECT 1
			FROM groups_users gu
			WHERE gu.group_id = g.id
			  AND gu.user_id = ?
		)
		AND g.is_private_chat = 0
	`

	args := []any{userID}

	if search != "" {
		query += `
			AND (
				g.name LIKE ?
				OR g.description LIKE ?
			)
		`

		searchPattern := "%" + search + "%"
		args = append(args, searchPattern, searchPattern)
	}

	query += `
		ORDER BY g.created_at
		LIMIT 12 OFFSET ?
	`

	args = append(args, offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var g models.Group

		if err := rows.Scan(
			&g.ID,
			&g.Title,
			&g.Description,
			&g.Avatar,
			&g.Count,
		); err != nil {
			return nil, err
		}

		groups = append(groups, g)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return groups, nil
}

func GetGroup(de *sql.DB, userID, groupID int) (models.Group, error) {
	var g models.Group

	err := de.QueryRow(`
	SELECT
		g.name,
		g.avatar,
		g.description,
		g.created_at,
		g.owner_id,
		(
			SELECT COUNT(*)
			FROM groups_users gu
			WHERE gu.group_id = g.id
		) AS member_count
	FROM groups g
	WHERE
		g.is_private_chat = 0
		AND g.id = ?
		AND EXISTS (
			SELECT 1
			FROM groups_users gu2
			WHERE gu2.group_id = g.id
			AND gu2.user_id = ?
		);
`, groupID, userID).Scan(
		&g.Title,
		&g.Avatar,
		&g.Description,
		&g.CreatedAt,
		&g.UserID,
		&g.Count,
	)

	return g, err
}

func SearchGroupMembers(db *sql.DB, groupID int, search string, offset int, limit int) ([]int, error) {
	rows, err := db.Query(`
		SELECT u.id
		FROM user u
		INNER JOIN groups_users gu ON gu.user_id = u.id
		WHERE gu.group_id = ?
		AND status = 1
		AND (
			u.first_name LIKE ?
			OR u.last_name LIKE ?
		)
		ORDER BY u.first_name, u.last_name
		LIMIT ? OFFSET ?
	`, groupID, "%"+search+"%", "%"+search+"%", limit, offset)

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

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ids, nil
}

func GroupExists(db *sql.DB, groupID int) (bool, error) {
	var exists bool

	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1
			FROM groups
			WHERE id = ?
			AND is_private_chat = 0
		)
	`, groupID).Scan(&exists)

	if err != nil {
		return false, err
	}

	return exists, nil
}

func UserIN(db *sql.DB, groupID, userID int) (bool, error) {
	var exists bool

	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1
			FROM groups_users
			WHERE group_id = ?
			AND user_id = ?
		)
	`, groupID, userID).Scan(&exists)

	if err != nil {
		return false, err
	}

	return exists, nil
}

func GetGroupData(db *sql.DB, groupID int) (models.Group, error) {
	var g models.Group
	err := db.QueryRow(`
		SELECT name, avatar FROM groups WHERE id = ?
	`, groupID).Scan(
		&g.Title,
		&g.Avatar,
	)

	return g, err
}

func SendGroupRequest(db *sql.DB, userID, groupID, code int) error {
	if code == -1 {
		_, err := db.Exec(`
		DELETE FROM groups_users
		WHERE user_id = ?
			AND group_id = ?
			AND status = 0
			AND invited_by IS NULL
	`, userID, groupID)
		return err
	}

	_, err := db.Exec(`
		INSERT INTO groups_users (group_id, user_id, status, invited_by) VALUES (?,?,0,NULL)
	`, groupID, userID)

	return err
}

func IsMember(db *sql.DB, groupID, userID int) (bool, error) {
	var exists bool

	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1
			FROM groups_users
			WHERE group_id = ?
				AND user_id = ?
				AND status = 1
		)
	`, groupID, userID).Scan(&exists)

	return exists, err
}

func HasJoinRequest(db *sql.DB, groupID, userID int) (bool, error) {
	var exists bool

	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1
			FROM groups_users
			WHERE group_id = ?
				AND user_id = ?
				AND status = 0
				AND invited_by IS NULL
		)
	`, groupID, userID).Scan(&exists)

	return exists, err
}

func GetGroupRequests(db *sql.DB, groupID, offset int) ([]models.UserRegistration, error) {
	users := make([]models.UserRegistration, 0)

	rows, err := db.Query(`
		SELECT
			u.id,
			u.first_name,
			u.last_name,
			p.avatar_path
		FROM groups_users gu
		JOIN user u ON u.id = gu.user_id
		JOIN profile p ON p.user_id = u.id
		WHERE gu.group_id = ?
			AND gu.status = 0
			AND gu.invited_by IS NULL
		ORDER BY u.first_name, u.last_name
		LIMIT 10 OFFSET ?
	`, groupID, offset)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var u models.UserRegistration

		if err := rows.Scan(
			&u.ID,
			&u.FirstName,
			&u.LastName,
			&u.Avatar,
		); err != nil {
			return nil, err
		}

		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func HandleGroupRequest(db *sql.DB, groupID, userID, code int) error {
	if code == 1 {
		_, err := db.Exec(`
			UPDATE groups_users
			SET status = 1
			WHERE group_id = ?
				AND user_id = ?
				AND status = 0
				AND invited_by IS NULL
		`, groupID, userID)

		return err
	}

	if code == -1 {
		_, err := db.Exec(`
			DELETE FROM groups_users
			WHERE group_id = ?
				AND user_id = ?
				AND status = 0
				AND invited_by IS NULL
		`, groupID, userID)

		return err
	}

	return fmt.Errorf("invalid request code")
}

func GetGroupOwner(db *sql.DB, groupID int) (int, error) {
	var id int
	err := db.QueryRow(`SELECT owner_id FROM groups WHERE id = ?`, groupID).Scan(&id)
	return id, err
}