package notifications

import (
	"database/sql"
	"social/internal/models"
)

func InsertNotification(db *sql.DB, n models.NewNotification) (int, error) {
	result, err := db.Exec(`
		INSERT INTO notifications (user_id, message)
		VALUES (?, ?)
	`, n.UserID, n.Message)

	if err != nil {
		return 0, err
	}

	notificationID, err := result.LastInsertId()

	if err != nil {
		return 0, err
	}

	_, err = db.Exec(`
		INSERT INTO notifications_types (
			notifications_id,
			message_user_id,
			post_id_tag,
			comment_id_tag,
			comment_reply_user_id,
			follow_request_user_id,
			follow_request_accept_user_id,
			follow_user_id,
			post_like_user_id,
			post_dislike_user_id,
			comment_like_user_id,
			comment_mention_user_id,
			post_mention_user_id,
			group_invite_user_id,
			group_join_user_id,
			group_accept_user_id,
			event_invite_user_id,
			event_response_user_id
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		notificationID,
		n.MessageUserID,
		n.PostIDTag,
		n.CommentIDTag,
		n.CommentReplyUserID,
		n.FollowRequestUserID,
		n.FollowRequestAcceptUserID,
		n.FollowUserID,
		n.PostLikeUserID,
		n.PostDislikeUserID,
		n.CommentLikeUserID,
		n.CommentMentionUserID,
		n.PostMentionUserID,
		n.GroupInviteUserID,
		n.GroupJoinUserID,
		n.GroupAcceptUserID,
		n.EventInviteUserID,
		n.EventResponseUserID,
	)

	if err != nil {
		return 0, err
	}

	return int(notificationID), nil
}

const actorExpression = `
	COALESCE(
		nt.message_user_id,
		nt.comment_reply_user_id,
		nt.follow_request_user_id,
		nt.follow_request_accept_user_id,
		nt.follow_user_id,
		nt.post_like_user_id,
		nt.post_dislike_user_id,
		nt.comment_like_user_id,
		nt.comment_mention_user_id,
		nt.post_mention_user_id,
		nt.group_invite_user_id,
		nt.group_join_user_id,
		nt.group_accept_user_id,
		nt.event_invite_user_id,
		nt.event_response_user_id
	)
`

const (
	duplicateWindow = "-10 minutes"
	burstWindow     = "-1 minute"
	burstLimit      = 5
)

func IsSpam(db *sql.DB, actorID int, n models.NewNotification) (bool, error) {
	if actorID <= 0 {
		return false, nil
	}

	var postID, commentID int

	if n.PostIDTag != nil {
		postID = *n.PostIDTag
	}

	if n.CommentIDTag != nil {
		commentID = *n.CommentIDTag
	}

	var duplicates int

	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM notifications n
		JOIN notifications_types nt
			ON nt.notifications_id = n.id
		WHERE n.user_id = ?
			AND n.message = ?
			AND `+actorExpression+` = ?
			AND IFNULL(nt.post_id_tag, 0) = ?
			AND IFNULL(nt.comment_id_tag, 0) = ?
			AND n.created_at >= datetime('now', ?)
	`, n.UserID, n.Message, actorID, postID, commentID, duplicateWindow).Scan(&duplicates)

	if err != nil {
		return false, err
	}

	if duplicates > 0 {
		return true, nil
	}

	var burst int

	err = db.QueryRow(`
		SELECT COUNT(*)
		FROM notifications n
		JOIN notifications_types nt
			ON nt.notifications_id = n.id
		WHERE n.user_id = ?
			AND `+actorExpression+` = ?
			AND n.created_at >= datetime('now', ?)
	`, n.UserID, actorID, burstWindow).Scan(&burst)

	if err != nil {
		return false, err
	}

	return burst >= burstLimit, nil
}

const excludedTypesClause = `
	nt.message_user_id IS NULL
	AND nt.group_invite_user_id IS NULL
	AND nt.group_join_user_id IS NULL
	AND nt.group_accept_user_id IS NULL
`

func GetUnreadCount(db *sql.DB, userID int) (int, error) {
	var count int

	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM notifications n
		LEFT JOIN notifications_types nt
			ON nt.notifications_id = n.id
		WHERE n.user_id = ?
			AND n.is_read = 0
			AND `+excludedTypesClause+`
	`, userID).Scan(&count)

	return count, err
}

func MarkAllRead(db *sql.DB, userID int) error {
	_, err := db.Exec(`
		UPDATE notifications
		SET is_read = 1
		WHERE user_id = ?
	`, userID)

	return err
}
