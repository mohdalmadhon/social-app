package posts

import (
	"database/sql"

	"social/database/dbutil"
	"social/internal/models"
)

// SearchPosts finds posts whose content contains the search text, limited to
// the posts the searching user is allowed to see:
//
//   - their own posts
//   - public posts (group id 0)
//   - followers-only posts (group id -1) of people they follow
//   - posts shared with a custom visibility group they belong to
func SearchPosts(db *sql.DB, userID int, search string, limit, offset int) ([]models.Post, error) {
	pattern := dbutil.LikePattern(search)

	rows, err := db.Query(`
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
			COALESCE(prx.value, 0),
			p.like_count,
			p.dislike_count,
			p.comment_count
		FROM posts p
		JOIN user u
			ON u.id = p.user_id
		LEFT JOIN profile pr
			ON pr.user_id = p.user_id
		LEFT JOIN post_reactions prx
			ON prx.post_id = p.id
			AND prx.user_id = ?
		LEFT JOIN user_posts_groups g
			ON g.id = p.group_id
			AND g.user_id = p.user_id
			AND p.group_id > 0
		WHERE p.content LIKE ? ESCAPE '\'
			AND (
				p.user_id = ?
				OR p.group_id = 0
				OR (
					p.group_id = -1
					AND EXISTS (
						SELECT 1
						FROM user_followers uf
						WHERE uf.follower_id = ?
							AND uf.target_id = p.user_id
							AND uf.status = 1
					)
				)
				OR (
					g.id IS NOT NULL
					AND (':' || g.users || ':') LIKE ('%:' || ? || ':%')
				)
			)
		ORDER BY p.created_at DESC, p.id DESC
		LIMIT ? OFFSET ?
	`, userID, pattern, userID, userID, userID, limit, offset)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	result := make([]models.Post, 0)
	tagsByPostIndex := make(map[int]string)

	for rows.Next() {
		var p models.Post
		var username sql.NullString
		var avatarPath sql.NullString
		var tags sql.NullString

		if err := rows.Scan(
			&p.Id,
			&p.UserId,
			&p.FirstName,
			&p.LastName,
			&username,
			&avatarPath,
			&p.Content,
			&p.ImagePath,
			&p.AllowComments,
			&p.Location,
			&p.CreatedAt,
			&p.GroupId,
			&tags,
			&p.ReactionValue,
			&p.LikeCount,
			&p.DisLikeCount,
			&p.CommentCount,
		); err != nil {
			return nil, err
		}

		if username.Valid {
			p.Username = &username.String
		}

		if avatarPath.Valid {
			p.AvatarPath = avatarPath.String
		}

		if tags.Valid {
			tagsByPostIndex[len(result)] = tags.String
		}

		result = append(result, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	rows.Close()

	if err := attachGroupOwnerNames(db, result); err != nil {
		return nil, err
	}

	if err := attachTaggedPeople(db, result, tagsByPostIndex); err != nil {
		return nil, err
	}

	return result, nil
}
