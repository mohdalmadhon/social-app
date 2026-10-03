package helpers

import "social/internal/models"

func GetNotificationType(notification models.NewNotification) string {
	if notification.MessageUserID != nil {
		return "message"
	}
	if notification.FollowRequestUserID != nil {
		return "follow"
	}
	if notification.FollowRequestAcceptUserID != nil {
		return "follow_accept"
	}
	if notification.FollowUserID != nil {
		return "follow"
	}
	if notification.PostLikeUserID != nil {
		return "like"
	}
	if notification.PostDislikeUserID != nil {
		return "dislike"
	}
	if notification.CommentLikeUserID != nil {
		return "comment like"
	}
	if notification.CommentMentionUserID != nil {
		return "comment mention"
	}
	if notification.PostMentionUserID != nil {
		return "post mention"
	}
	if notification.PostIDTag != nil {
		return "post tag"
	}
	if notification.CommentIDTag != nil {
		return "comment tag"
	}
	if notification.CommentReplyUserID != nil {
		return "comment reply"
	}
	if notification.GroupInviteUserID != nil {
		return "group invite"
	}
	if notification.GroupJoinUserID != nil {
		return "group join"
	}
	if notification.GroupAcceptUserID != nil {
		return "group accept"
	}
	if notification.EventInviteUserID != nil {
		return "event invite"
	}
	if notification.EventResponseUserID != nil {
		return "event response"
	}

	return ""
}

func GetNotificationPreferenceType(notification models.NewNotification) string {
	switch {
	case notification.FollowUserID != nil:
		return "follow"
	case notification.PostLikeUserID != nil, notification.PostDislikeUserID != nil:
		return "post_reaction"
	case notification.CommentReplyUserID != nil:
		return "comment"
	case notification.CommentLikeUserID != nil:
		return "comment_like"
	case notification.CommentMentionUserID != nil, notification.PostMentionUserID != nil:
		return "mention"
	case notification.EventInviteUserID != nil, notification.EventResponseUserID != nil:
		return "event"
	}

	return ""
}
