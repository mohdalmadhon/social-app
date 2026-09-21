package posts

import (
	"database/sql"
	"social/internal/models"
)

func GetSinglePost(db *sql.DB, postID, userID int) (models.Post, error) {
	var post models.Post
	var username sql.NullString
	var avatarPath sql.NullString
	var tags sql.NullString

	err := db.QueryRow(`
		SELECT
			p.id,
			p.user_id,
			u.first_name,
			u.last_name,
			u.username,
			pr.avatar_path,
			p.content,
			p.image_path,
			p.allow_comments,
			p.location,
			p.created_at,
			p.group_id,
			p.tags,
			COALESCE((
				SELECT r.value
				FROM post_reactions r
				WHERE r.post_id = p.id
				AND r.user_id = ?
			), 0),
			(
				SELECT COUNT(*)
				FROM post_reactions r
				WHERE r.post_id = p.id
				AND r.value = 1
			),
			(
				SELECT COUNT(*)
				FROM post_reactions r
				WHERE r.post_id = p.id
				AND r.value = -1
			),
			(
				SELECT COUNT(*)
				FROM comments c
				WHERE c.post_id = p.id
			)
		FROM posts p
		JOIN user u ON u.id = p.user_id
		LEFT JOIN profile pr ON pr.user_id = p.user_id
		WHERE p.id = ?
	`, userID, postID).Scan(
		&post.Id,
		&post.UserId,
		&post.FirstName,
		&post.LastName,
		&username,
		&avatarPath,
		&post.Content,
		&post.ImagePath,
		&post.AllowComments,
		&post.Location,
		&post.CreatedAt,
		&post.GroupId,
		&tags,
		&post.ReactionValue,
		&post.LikeCount,
		&post.DisLikeCount,
		&post.CommentCount,
	)

	if err != nil {
		return post, err
	}

	if username.Valid {
		post.Username = &username.String
	}

	if avatarPath.Valid {
		post.AvatarPath = avatarPath.String
	}

	post.TaggedPeople = []models.TaggedPerson{}

	list := []models.Post{post}

	tagsByPostIndex := map[int]string{}

	if tags.Valid && tags.String != "" {
		tagsByPostIndex[0] = tags.String
	}

	if err := attachGroupOwnerNames(db, list); err != nil {
		return post, err
	}

	if err := attachTaggedPeople(db, list, tagsByPostIndex); err != nil {
		return post, err
	}

	result := list[0]

	if result.TaggedPeople == nil {
		result.TaggedPeople = []models.TaggedPerson{}
	}

	return result, nil
}
