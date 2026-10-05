package dbutil

import (
	"database/sql"
	"fmt"
	"strings"

	"social/internal/models"
)

const (
	PostTagsTable      = "post_user_tags"
	GroupPostTagsTable = "group_post_user_tags"
)

type Execer interface {
	Exec(query string, args ...any) (sql.Result, error)
}

func checkTagsTable(table string) error {
	if table != PostTagsTable && table != GroupPostTagsTable {
		return fmt.Errorf("invalid tags table: %s", table)
	}
	return nil
}

func InsertTags(exec Execer, table string, postID int, userIDs []int) error {
	if err := checkTagsTable(table); err != nil {
		return err
	}

	seen := make(map[int]bool)

	for _, userID := range userIDs {
		if userID <= 0 || seen[userID] {
			continue
		}

		seen[userID] = true

		_, err := exec.Exec(`
			INSERT OR IGNORE INTO `+table+` (user_id, post_id)
			SELECT ?, ?
			WHERE EXISTS (SELECT 1 FROM user WHERE id = ?)
		`, userID, postID, userID)

		if err != nil {
			return err
		}
	}

	return nil
}

func AttachTaggedPeople(db *sql.DB, table string, posts []models.Post) error {
	if err := checkTagsTable(table); err != nil {
		return err
	}

	indexesByPostID := make(map[int][]int)

	for i := range posts {
		posts[i].TaggedPeople = []models.TaggedPerson{}
		indexesByPostID[posts[i].Id] = append(indexesByPostID[posts[i].Id], i)
	}

	if len(indexesByPostID) == 0 {
		return nil
	}

	placeholders := make([]string, 0, len(indexesByPostID))
	args := make([]interface{}, 0, len(indexesByPostID))

	for postID := range indexesByPostID {
		placeholders = append(placeholders, "?")
		args = append(args, postID)
	}

	rows, err := db.Query(`
		SELECT t.post_id, u.id, u.first_name, u.last_name, pr.avatar_path
		FROM `+table+` t
		JOIN user u ON u.id = t.user_id
		LEFT JOIN profile pr ON pr.user_id = u.id
		WHERE t.post_id IN (`+strings.Join(placeholders, ",")+`)
		ORDER BY t.rowid
	`, args...)

	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		var postID int
		var person models.TaggedPerson
		var avatarPath sql.NullString

		if err := rows.Scan(
			&postID,
			&person.Id,
			&person.FirstName,
			&person.LastName,
			&avatarPath,
		); err != nil {
			return err
		}

		if avatarPath.Valid {
			person.AvatarPath = avatarPath.String
		}

		for _, i := range indexesByPostID[postID] {
			posts[i].TaggedPeople = append(posts[i].TaggedPeople, person)
		}
	}

	return rows.Err()
}
