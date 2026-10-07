package groups

import (
	"database/sql"
	"social/internal/models"
)

func UserInPostGroup(db *sql.DB, postID, userID int) (bool, error) {
	var allowed bool

	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1
			FROM group_posts gp
			JOIN groups_users gu
				ON gu.group_id = gp.group_id
			WHERE gp.id = ?
			AND gu.user_id = ?
			AND gu.status = 1
		)
	`, postID, userID).Scan(&allowed)

	if err != nil {
		return false, err
	}

	return allowed, nil
}

func UserOwnsGroupComment(db *sql.DB, commentID, userID int) (bool, error) {
	var owns bool

	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1
			FROM group_comments
			WHERE id = ?
			AND user_id = ?
		)
	`, commentID, userID).Scan(&owns)

	if err != nil {
		return false, err
	}

	return owns, nil
}

func GroupCommentPostID(db *sql.DB, commentID int) (int, error) {
	var postID int

	err := db.QueryRow(`
		SELECT post_id
		FROM group_comments
		WHERE id = ?
	`, commentID).Scan(&postID)

	if err != nil {
		return 0, err
	}

	return postID, nil
}

func InsertGroupComment(db *sql.DB, comment models.GroupComment) (models.GroupComment, error) {
	result, err := db.Exec(`
		INSERT INTO group_comments (user_id, post_id, content, reply_to, image_path)
		VALUES (?, ?, ?, ?, NULLIF(?, ''))
	`, comment.User.ID, comment.GroupPostID, comment.Content, comment.ReplyTo, comment.ImagePath)

	if err != nil {
		return comment, err
	}

	id, err := result.LastInsertId()

	if err != nil {
		return comment, err
	}

	comment.ID = int(id)

	err = db.QueryRow(`
		SELECT created_at
		FROM group_comments
		WHERE id = ?
	`, comment.ID).Scan(&comment.CreatedAt)

	if err != nil {
		return comment, err
	}

	err = db.QueryRow(`
		SELECT
			u.first_name,
			u.last_name,
			COALESCE(p.avatar_path, '')
		FROM user u
		LEFT JOIN profile p
			ON p.user_id = u.id
		WHERE u.id = ?
	`, comment.User.ID).Scan(
		&comment.User.FirstName,
		&comment.User.LastName,
		&comment.User.Avatar,
	)

	if err != nil {
		return comment, err
	}

	comment.Votes = 0
	comment.Replies = 0

	return comment, nil
}

func GetGroupComments(db *sql.DB, postID, replyTo, limit, offset int, orderBy string) ([]models.GroupComment, error) {
	queryOrderBy := "c.created_at DESC"

	if orderBy == "popular" {
		queryOrderBy = "c.votes DESC, c.created_at DESC"
	}

	selection := `
		SELECT
			u.id,
			u.first_name,
			u.last_name,
			COALESCE(p.avatar_path, ''),
			c.id,
			c.content,
			COALESCE(c.image_path, ''),
			c.post_id,
			c.created_at,
			c.reply_to,
			c.votes,
			(
				SELECT COUNT(*)
				FROM group_comments r
				WHERE r.reply_to = c.id
			)
		FROM group_comments c
		JOIN user u ON u.id = c.user_id
		LEFT JOIN profile p ON p.user_id = u.id
	`

	var rows *sql.Rows
	var err error

	if replyTo == 0 {
		rows, err = db.Query(selection+`
			WHERE c.post_id = ?
			AND c.reply_to IS NULL
			ORDER BY `+queryOrderBy+`
			LIMIT ?
			OFFSET ?
		`, postID, limit, offset)
	} else {
		rows, err = db.Query(selection+`
			WHERE c.post_id = ?
			AND c.reply_to = ?
			ORDER BY `+queryOrderBy+`
			LIMIT ?
			OFFSET ?
		`, postID, replyTo, limit, offset)
	}

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var comments []models.GroupComment

	for rows.Next() {
		var comment models.GroupComment

		err := rows.Scan(
			&comment.User.ID,
			&comment.User.FirstName,
			&comment.User.LastName,
			&comment.User.Avatar,
			&comment.ID,
			&comment.Content,
			&comment.ImagePath,
			&comment.GroupPostID,
			&comment.CreatedAt,
			&comment.ReplyTo,
			&comment.Votes,
			&comment.Replies,
		)

		if err != nil {
			return nil, err
		}

		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}

func CountGroupComments(db *sql.DB, postID int) (int, error) {
	var count int

	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM group_comments
		WHERE post_id = ?
	`, postID).Scan(&count)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func DeleteGroupComment(db *sql.DB, commentID, userID int) ([]string, error) {
	imageRows, err := db.Query(`
		WITH RECURSIVE tree(id) AS (
			SELECT id FROM group_comments WHERE id = ? AND (user_id = ? OR post_id IN (SELECT id FROM group_posts WHERE user_id = ?))
			UNION ALL
			SELECT c.id FROM group_comments c JOIN tree t ON c.reply_to = t.id
		)
		SELECT image_path FROM group_comments
		WHERE id IN (SELECT id FROM tree)
		AND image_path IS NOT NULL
		AND image_path != ''
	`, commentID, userID, userID)

	if err != nil {
		return nil, err
	}

	var imagePaths []string

	for imageRows.Next() {
		var imagePath string

		if err := imageRows.Scan(&imagePath); err != nil {
			imageRows.Close()
			return nil, err
		}

		imagePaths = append(imagePaths, imagePath)
	}

	if err := imageRows.Err(); err != nil {
		imageRows.Close()
		return nil, err
	}

	imageRows.Close()

	result, err := db.Exec(`
		DELETE FROM group_comments
		WHERE id = ?
		AND (user_id = ? OR post_id IN (SELECT id FROM group_posts WHERE user_id = ?))
	`, commentID, userID, userID)

	if err != nil {
		return nil, err
	}

	rows, err := result.RowsAffected()

	if err != nil {
		return nil, err
	}

	if rows == 0 {
		return nil, sql.ErrNoRows
	}

	return imagePaths, nil
}

func VoteGroupComment(db *sql.DB, commentID, userID, vote int) error {
	tx, err := db.Begin()

	if err != nil {
		return err
	}

	defer tx.Rollback()

	var currentVote int

	err = tx.QueryRow(`
		SELECT count
		FROM group_comment_votes
		WHERE comment_id = ?
		AND user_id = ?
	`, commentID, userID).Scan(&currentVote)

	if err == sql.ErrNoRows {
		_, err = tx.Exec(`
			INSERT INTO group_comment_votes
			(user_id, comment_id, count)
			VALUES (?, ?, ?)
		`, userID, commentID, vote)

		if err != nil {
			return err
		}

		_, err = tx.Exec(`
			UPDATE group_comments
			SET votes = votes + ?
			WHERE id = ?
		`, vote, commentID)

		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		if currentVote == vote {
			_, err = tx.Exec(`
				DELETE FROM group_comment_votes
				WHERE comment_id = ?
				AND user_id = ?
			`, commentID, userID)

			if err != nil {
				return err
			}

			_, err = tx.Exec(`
				UPDATE group_comments
				SET votes = votes - ?
				WHERE id = ?
			`, vote, commentID)

			if err != nil {
				return err
			}
		} else {
			_, err = tx.Exec(`
				UPDATE group_comment_votes
				SET count = ?
				WHERE comment_id = ?
				AND user_id = ?
			`, vote, commentID, userID)

			if err != nil {
				return err
			}

			_, err = tx.Exec(`
				UPDATE group_comments
				SET votes = votes + ?
				WHERE id = ?
			`, vote*2, commentID)

			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}
