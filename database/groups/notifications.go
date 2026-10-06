package groups

import "database/sql"

func GetGroupPostMeta(db *sql.DB, postID int) (int, int, error) {
	var ownerID, groupID int

	err := db.QueryRow(`
		SELECT user_id, COALESCE(group_id, 0)
		FROM group_posts
		WHERE id = ?
	`, postID).Scan(&ownerID, &groupID)

	return ownerID, groupID, err
}

func GetGroupCommentMeta(db *sql.DB, commentID int) (int, int, int, error) {
	var ownerID, postID, groupID int

	err := db.QueryRow(`
		SELECT gc.user_id, gc.post_id, COALESCE(gp.group_id, 0)
		FROM group_comments gc
		JOIN group_posts gp
			ON gp.id = gc.post_id
		WHERE gc.id = ?
	`, commentID).Scan(&ownerID, &postID, &groupID)

	return ownerID, postID, groupID, err
}

func GetGroupPostReactionValue(db *sql.DB, postID, userID int) (int, error) {
	var value int

	err := db.QueryRow(`
		SELECT value
		FROM group_post_reactions
		WHERE post_id = ?
			AND user_id = ?
	`, postID, userID).Scan(&value)

	if err == sql.ErrNoRows {
		return 0, nil
	}

	return value, err
}

func GetGroupCommentVote(db *sql.DB, commentID, userID int) (int, error) {
	var vote int

	err := db.QueryRow(`
		SELECT count
		FROM group_comment_votes
		WHERE comment_id = ?
			AND user_id = ?
	`, commentID, userID).Scan(&vote)

	if err == sql.ErrNoRows {
		return 0, nil
	}

	return vote, err
}

func GetPendingInviter(db *sql.DB, groupID, userID int) (int, error) {
	var inviterID sql.NullInt64

	err := db.QueryRow(`
		SELECT invited_by
		FROM groups_users
		WHERE group_id = ?
			AND user_id = ?
			AND status = 0
			AND invited_by IS NOT NULL
	`, groupID, userID).Scan(&inviterID)

	if err == sql.ErrNoRows {
		return 0, nil
	}

	if err != nil {
		return 0, err
	}

	return int(inviterID.Int64), nil
}
