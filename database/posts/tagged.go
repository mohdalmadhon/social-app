package posts

import (
	"database/sql"
	"social/internal/models"
)

func GetTaggedPosts(db *sql.DB, targetID, offset, limit int) ([]models.Post, error) {
	rows, err := db.Query(`
		SELECT
			p.id,
			u.id,
			u.first_name,
			u.last_name,
			p.content,
			p.image_path,
			p.allow_comments,
			p.location,
			p.group_id,
			p.created_at,
			p.like_count,
			p.dislike_count,
			p.comment_count,
			p.public,
			p.private,
			COALESCE(pr.avatar_path, '')
		FROM posts p
		JOIN user u
			ON u.id = p.user_id
		LEFT JOIN profile pr
			ON pr.user_id = u.id
		WHERE p.id IN (
			SELECT post_id
			FROM post_user_tags
			WHERE user_id = ?
		)
		ORDER BY p.created_at DESC, p.id DESC
		LIMIT ?
		OFFSET ?
	`, targetID, limit, offset)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var posts []models.Post

	for rows.Next() {
		var p models.Post

		err := rows.Scan(
			&p.Id,
			&p.UserId,
			&p.FirstName,
			&p.LastName,
			&p.Content,
			&p.ImagePath,
			&p.AllowComments,
			&p.Location,
			&p.GroupId,
			&p.CreatedAt,
			&p.LikeCount,
			&p.DisLikeCount,
			&p.CommentCount,
			&p.Public,
			&p.Private,
			&p.AvatarPath,
		)

		if err != nil {
			return nil, err
		}

		posts = append(posts, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range posts {
		p := &posts[i]

		if p.GroupId != nil && *p.GroupId > 0 {
			var groupName string

			err := db.QueryRow(`
				SELECT name
				FROM user_posts_groups
				WHERE id = ?
			`, *p.GroupId).Scan(&groupName)

			if err != nil && err != sql.ErrNoRows {
				return nil, err
			}

			p.GroupName = groupName
		}
	}

	if err := attachTaggedPeople(db, &posts); err != nil {
		return nil, err
	}

	return posts, nil
}

func FilterTaggedPosts(db *sql.DB, posts []models.Post, viewerID, targetID int) ([]models.Post, error) {
	visiblePosts := []models.Post{}

	for _, p := range posts {
		if viewerID == targetID || viewerID == p.UserId || isTaggedUser(p, viewerID) {
			visiblePosts = append(visiblePosts, p)
			continue
		}

		allowed, err := FilterPosts(db, &[]models.Post{p}, viewerID, p.UserId)
		if err != nil {
			return nil, err
		}

		visiblePosts = append(visiblePosts, allowed...)
	}

	return visiblePosts, nil
}

func isTaggedUser(p models.Post, userID int) bool {
	for _, tagged := range p.TaggedPeople {
		if tagged.Id == userID {
			return true
		}
	}

	return false
}

func IsPrivateProfile(db *sql.DB, userID int) (bool, error) {
	var isPrivate int

	err := db.QueryRow(`
		SELECT COALESCE(is_private, 0)
		FROM profile
		WHERE user_id = ?
	`, userID).Scan(&isPrivate)

	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}

		return false, err
	}

	return isPrivate == 1, nil
}
