package groups

import (
	"database/sql"
	"fmt"
	"social/database/dbutil"
	"social/database/users"
	"social/internal/models"
)

func GetPosts(db *sql.DB, groupID, userID, offset int) ([]models.Post, error) {
	var posts []models.Post
	rows, err := db.Query(`
		SELECT id, content, user_id, image_path, location, created_at FROM group_posts
		WHERE 
			group_id = ?
		AND EXISTS (
			SELECT 1 FROM groups_users WHERE user_id = ? AND group_id = ?
		)
		ORDER BY created_at DESC
		LIMIT 10
		OFFSET ?
	`, groupID, userID, groupID, offset)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var p models.Post
		err := rows.Scan(
			&p.Id,
			&p.Content,
			&p.UserId,
			&p.ImagePath,
			&p.Location,
			&p.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		posts = append(posts, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	rows.Close()

	if err := dbutil.AttachTaggedPeople(db, dbutil.GroupPostTagsTable, posts); err != nil {
		return nil, err
	}

	return posts, nil
}

func AddPost(db *sql.DB, post models.Post, groupID int, taggedPeople []int) (int, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}

	defer tx.Rollback()

	result, err := tx.Exec(`
	INSERT INTO group_posts (
		user_id,
		content,
		image_path,
		location,
		group_id
	)
	SELECT ?, ?, ?, ?, ?
	WHERE EXISTS (
		SELECT 1
		FROM groups_users
		WHERE group_id = ?
		AND user_id = ?
		AND status = 1
	)
`,
		post.UserId,
		post.Content,
		post.ImagePath,
		post.Location,
		groupID,
		groupID,
		post.UserId,
	)

	if err != nil {
		return 0, err
	}

	rows, err := result.RowsAffected()

	if err != nil {
		return 0, err
	}

	if rows == 0 {
		return 0, fmt.Errorf("user is not a member of the group")
	}

	id, err := result.LastInsertId()

	if err != nil {
		return 0, err
	}

	if err := dbutil.InsertTags(tx, dbutil.GroupPostTagsTable, int(id), taggedPeople); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return int(id), nil
}

func GetGroupPost(db *sql.DB, postID int, userID int) (models.Post, error) {
	var post models.Post

	var reaction sql.NullInt64
	var groupID sql.NullInt64

	query := `
		SELECT
			gp.id,
			gp.user_id,
			gp.content,
			gp.image_path,
			gp.location,
			gp.group_id,
			gp.created_at,
			g.name,

			COALESCE(SUM(CASE WHEN pr.value = 1 THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN pr.value = -1 THEN 1 ELSE 0 END), 0),
			(
				SELECT COUNT(*)
				FROM group_comments c
				WHERE c.post_id = gp.id
			),

			(
				SELECT pr2.value
				FROM group_post_reactions pr2
				WHERE pr2.post_id = gp.id
				AND pr2.user_id = ?
			)

		FROM group_posts gp

		LEFT JOIN groups g
			ON g.id = gp.group_id

		LEFT JOIN group_post_reactions pr
			ON pr.post_id = gp.id

		WHERE gp.id = ?

		GROUP BY
			gp.id,
			gp.user_id,
			gp.content,
			gp.image_path,
			gp.location,
			gp.group_id,
			gp.created_at,
			g.name
	`

	err := db.QueryRow(query, userID, postID).Scan(
		&post.Id,
		&post.UserId,
		&post.Content,
		&post.ImagePath,
		&post.Location,
		&groupID,
		&post.CreatedAt,
		&post.GroupName,
		&post.LikeCount,
		&post.DisLikeCount,
		&post.CommentCount,
		&reaction,
	)

	if err != nil {
		return post, err
	}

	if groupID.Valid {
		id := int(groupID.Int64)
		post.GroupId = &id
	}

	if reaction.Valid {
		post.ReactionValue = int(reaction.Int64)
	}

	user, err := users.GetUserSimpleData(db, post.UserId)
	if err != nil {
		return post, err
	}

	post.FirstName = user.FirstName
	post.LastName = user.LastName
	post.Username = &user.UserName
	post.AvatarPath = user.Avatar

	post.AllowComments = true

	list := []models.Post{post}

	if err := dbutil.AttachTaggedPeople(db, dbutil.GroupPostTagsTable, list); err != nil {
		return post, err
	}

	return list[0], nil
}

func InsertReaction(db *sql.DB, reaction models.Reaction) error {

	var currentValue int

	err := db.QueryRow(`
		SELECT value
		FROM group_post_reactions
		WHERE user_id = ? AND post_id = ?
	`, reaction.UserID, reaction.PostID).Scan(&currentValue)

	if err != nil && err != sql.ErrNoRows {
		return err
	}
	
	if reaction.Value == currentValue {
		_, err := db.Exec(`
			DELETE FROM group_post_reactions
			WHERE post_id = ? AND user_id = ?
		`, reaction.PostID, reaction.UserID)

		return err
	}

	if err == sql.ErrNoRows {
		_, err := db.Exec(`
			INSERT INTO group_post_reactions (post_id, user_id, value)
			VALUES (?, ?, ?)
		`, reaction.PostID, reaction.UserID, reaction.Value)

		return err
	}

	_, err = db.Exec(`
		UPDATE group_post_reactions
		SET value = ?
		WHERE user_id = ? AND post_id = ?
	`, reaction.Value, reaction.UserID, reaction.PostID)

	return err
}
