package posts

import "database/sql"

func GetPostOwnerID(db *sql.DB, postID int) (int, error) {
	var ID int
	err := db.QueryRow(`
		SELECT user_id FROM posts WHERE id = ?
	`, postID).Scan(&ID)

	return ID, err
}

// GetPostImage returns the image path of a post ("" when it has none).
func GetPostImage(db *sql.DB, postID int) (string, error) {
	var imagePath sql.NullString

	err := db.QueryRow(`
		SELECT image_path FROM posts WHERE id = ?
	`, postID).Scan(&imagePath)

	return imagePath.String, err
}

// GetCommentOwner returns the author of a comment and the post it belongs to.
func GetCommentOwner(db *sql.DB, commentID int) (int, int, error) {
	var ownerID, postID int

	err := db.QueryRow(`
		SELECT user_id, post_id FROM comments WHERE id = ?
	`, commentID).Scan(&ownerID, &postID)

	return ownerID, postID, err
}

// GetUserCommentVote returns the vote (1 / -1) a user currently has on a
// comment, or 0 when they have none.
func GetUserCommentVote(db *sql.DB, commentID, userID int) (int, error) {
	var vote int

	err := db.QueryRow(`
		SELECT count FROM comment_votes
		WHERE comment_id = ? AND user_id = ?
	`, commentID, userID).Scan(&vote)

	if err == sql.ErrNoRows {
		return 0, nil
	}

	return vote, err
}
