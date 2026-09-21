package groups

import (
	"database/sql"
	"social/internal/models"
)

func GetGroupFeedPosts(db *sql.DB, groupID, userID, limit, offset int) ([]models.Post, error) {
	rows, err := db.Query(`
		SELECT
			gp.id,
			gp.user_id,
			gp.content,
			gp.image_path,
			gp.location,
			gp.group_id,
			gp.created_at,
			u.first_name,
			u.last_name,
			COALESCE(u.username, ''),
			COALESCE(p.avatar_path, ''),
			(
				SELECT COUNT(*)
				FROM group_post_reactions pr
				WHERE pr.post_id = gp.id
				AND pr.value = 1
			),
			(
				SELECT COUNT(*)
				FROM group_post_reactions pr
				WHERE pr.post_id = gp.id
				AND pr.value = -1
			),
			(
				SELECT COUNT(*)
				FROM group_comments c
				WHERE c.post_id = gp.id
			),
			COALESCE((
				SELECT pr2.value
				FROM group_post_reactions pr2
				WHERE pr2.post_id = gp.id
				AND pr2.user_id = ?
			), 0)
		FROM group_posts gp
		JOIN user u
			ON u.id = gp.user_id
		LEFT JOIN profile p
			ON p.user_id = gp.user_id
		WHERE gp.group_id = ?
		ORDER BY gp.created_at DESC
		LIMIT ?
		OFFSET ?
	`, userID, groupID, limit, offset)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	posts := []models.Post{}

	for rows.Next() {
		var post models.Post
		var username string
		var groupIDValue sql.NullInt64

		err := rows.Scan(
			&post.Id,
			&post.UserId,
			&post.Content,
			&post.ImagePath,
			&post.Location,
			&groupIDValue,
			&post.CreatedAt,
			&post.FirstName,
			&post.LastName,
			&username,
			&post.AvatarPath,
			&post.LikeCount,
			&post.DisLikeCount,
			&post.CommentCount,
			&post.ReactionValue,
		)

		if err != nil {
			return nil, err
		}

		if groupIDValue.Valid {
			id := int(groupIDValue.Int64)
			post.GroupId = &id
		}

		post.Username = &username
		post.AllowComments = true
		post.TaggedPeople = []models.TaggedPerson{}

		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}
