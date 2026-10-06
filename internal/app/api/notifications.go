package api

import (
	"database/sql"
	"log"
	"net/http"
	"social/database/notifications"
	"social/internal/helpers"
	"strconv"
)

func (app *App) GetNotification(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)

	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	offset := 0

	if value := r.URL.Query().Get("offset"); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	rows, err := app.DB.Query(`
		SELECT
			n.id,
			n.message,
			n.created_at,
			n.is_read,
			nt.notifications_id,
			nt.post_id_tag,
			nt.comment_id_tag,
			nt.comment_reply_user_id,
			nt.follow_request_user_id,
			nt.follow_request_accept_user_id,
			nt.follow_user_id,
			nt.post_like_user_id,
			nt.post_dislike_user_id,
			nt.comment_like_user_id,
			nt.comment_mention_user_id,
			nt.post_mention_user_id,
			nt.event_invite_user_id,
			nt.event_response_user_id,
			nt.group_invite_user_id,
			nt.group_join_user_id,
			nt.group_accept_user_id,
			nt.group_id,
			nt.event_id,
			ge.title,
			ge.event_time,
			actor.id,
			actor.first_name,
			actor.last_name,
			actor_profile.avatar_path,
			p.id,
			p.content,
			p.image_path,
			c.content,
			c.votes,
			gr.name,
			gr.avatar
		FROM notifications n
		JOIN notifications_types nt
			ON nt.notifications_id = n.id
		LEFT JOIN user actor
			ON actor.id = COALESCE(
				nt.comment_reply_user_id,
				nt.follow_request_user_id,
				nt.follow_request_accept_user_id,
				nt.follow_user_id,
				nt.post_like_user_id,
				nt.post_dislike_user_id,
				nt.comment_like_user_id,
				nt.comment_mention_user_id,
				nt.post_mention_user_id,
				nt.event_invite_user_id,
				nt.event_response_user_id,
				nt.group_invite_user_id,
				nt.group_join_user_id,
				nt.group_accept_user_id
			)
		LEFT JOIN profile actor_profile
			ON actor_profile.user_id = actor.id
		LEFT JOIN posts p
			ON p.id = nt.post_id_tag
		LEFT JOIN comments c
			ON c.id = nt.comment_id_tag
		LEFT JOIN groups gr
			ON gr.id = nt.group_id
		LEFT JOIN group_events ge
			ON ge.id = nt.event_id
		WHERE n.user_id = ?
			AND nt.message_user_id IS NULL
		ORDER BY n.created_at DESC, n.id DESC
		LIMIT 20 OFFSET ?
	`, userID, offset)

	if err != nil {
		log.Println(err)
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not get notifications",
		})
		return
	}
	defer rows.Close()

	notifications := []map[string]any{}

	for rows.Next() {
		var (
			id                        int
			message                   string
			createdAt                 string
			isRead                    int
			notificationsID           int
			postIDTag                 sql.NullInt64
			commentIDTag              sql.NullInt64
			commentReplyUserID        sql.NullInt64
			followRequestUserID       sql.NullInt64
			followRequestAcceptUserID sql.NullInt64
			followUserID              sql.NullInt64
			postLikeUserID            sql.NullInt64
			postDislikeUserID         sql.NullInt64
			commentLikeUserID         sql.NullInt64
			commentMentionUserID      sql.NullInt64
			postMentionUserID         sql.NullInt64
			eventInviteUserID         sql.NullInt64
			eventResponseUserID       sql.NullInt64
			groupInviteUserID         sql.NullInt64
			groupJoinUserID           sql.NullInt64
			groupAcceptUserID         sql.NullInt64
			groupID                   sql.NullInt64
			eventID                   sql.NullInt64
			eventTitle                sql.NullString
			eventTime                 sql.NullString
			actorID                   sql.NullInt64
			actorFirstName            sql.NullString
			actorLastName             sql.NullString
			actorAvatarPath           sql.NullString
			postID                    sql.NullInt64
			postContent               sql.NullString
			postImagePath             sql.NullString
			commentContent            sql.NullString
			commentVotes              sql.NullInt64
			groupName                 sql.NullString
			groupAvatar               sql.NullString
		)

		err := rows.Scan(
			&id,
			&message,
			&createdAt,
			&isRead,
			&notificationsID,
			&postIDTag,
			&commentIDTag,
			&commentReplyUserID,
			&followRequestUserID,
			&followRequestAcceptUserID,
			&followUserID,
			&postLikeUserID,
			&postDislikeUserID,
			&commentLikeUserID,
			&commentMentionUserID,
			&postMentionUserID,
			&eventInviteUserID,
			&eventResponseUserID,
			&groupInviteUserID,
			&groupJoinUserID,
			&groupAcceptUserID,
			&groupID,
			&eventID,
			&eventTitle,
			&eventTime,
			&actorID,
			&actorFirstName,
			&actorLastName,
			&actorAvatarPath,
			&postID,
			&postContent,
			&postImagePath,
			&commentContent,
			&commentVotes,
			&groupName,
			&groupAvatar,
		)

		if err != nil {
			log.Println(err)
			helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
				"status":  false,
				"message": "could not read notifications",
			})
			return
		}

		notification := map[string]any{
			"id":         id,
			"message":    message,
			"created_at": createdAt,
			"is_read":    isRead == 1,
		}

		if postIDTag.Valid {
			notification["post_id"] = postIDTag.Int64
		}

		if commentReplyUserID.Valid {
			notification["comment_reply_user_id"] = commentReplyUserID.Int64
		}

		if followRequestUserID.Valid {
			notification["follow_request_user_id"] = followRequestUserID.Int64
		}

		if followRequestAcceptUserID.Valid {
			notification["follow_request_accept_user_id"] = followRequestAcceptUserID.Int64
		}

		if followUserID.Valid {
			notification["follow_user_id"] = followUserID.Int64
		}

		if postLikeUserID.Valid {
			notification["post_like_user_id"] = postLikeUserID.Int64
		}

		if postDislikeUserID.Valid {
			notification["post_dislike_user_id"] = postDislikeUserID.Int64
		}

		if commentLikeUserID.Valid {
			notification["comment_like_user_id"] = commentLikeUserID.Int64
		}

		if commentMentionUserID.Valid {
			notification["comment_mention_user_id"] = commentMentionUserID.Int64
		}

		if postMentionUserID.Valid {
			notification["post_mention_user_id"] = postMentionUserID.Int64
		}

		if eventInviteUserID.Valid {
			notification["event_invite_user_id"] = eventInviteUserID.Int64
		}

		if eventResponseUserID.Valid {
			notification["event_response_user_id"] = eventResponseUserID.Int64
		}

		if groupInviteUserID.Valid {
			notification["group_invite_user_id"] = groupInviteUserID.Int64
		}

		if groupJoinUserID.Valid {
			notification["group_join_user_id"] = groupJoinUserID.Int64
		}

		if groupAcceptUserID.Valid {
			notification["group_accept_user_id"] = groupAcceptUserID.Int64
		}

		if eventID.Valid {
			notification["event"] = map[string]any{
				"id":        eventID.Int64,
				"title":     eventTitle.String,
				"eventTime": eventTime.String,
			}
		}

		if groupID.Valid {
			notification["group"] = map[string]any{
				"id":     groupID.Int64,
				"name":   groupName.String,
				"avatar": groupAvatar.String,
			}
		}

		if actorID.Valid {
			notification["actor"] = map[string]any{
				"id":         actorID.Int64,
				"firstName":  actorFirstName.String,
				"lastName":   actorLastName.String,
				"avatarPath": actorAvatarPath.String,
			}
		}

		if postID.Valid {
			notification["post"] = map[string]any{
				"id":        postID.Int64,
				"content":   postContent.String,
				"imagePath": postImagePath.String,
			}
		}

		if commentIDTag.Valid && commentContent.Valid {
			notification["comment"] = commentContent.String
			notification["comment_likes"] = commentVotes.Int64
		}

		notifications = append(notifications, notification)
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not read notifications",
		})
		return
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status":        true,
		"notifications": notifications,
		"offset":        offset,
		"limit":         20,
		"hasMore":       len(notifications) == 20,
	})
}

func (app *App) GetUnreadNotificationCount(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)

	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	count, err := notifications.GetUnreadCount(app.DB, userID)

	if err != nil {
		log.Println(err)
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not get unread count",
		})
		return
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status": true,
		"count":  count,
	})
}

func (app *App) MarkNotificationsRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)

	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	if err := notifications.MarkAllRead(app.DB, userID); err != nil {
		log.Println(err)
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not mark notifications read",
		})
		return
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status": true,
	})
}

func (app *App) MarkNotificationRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)

	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	notificationID, err := strconv.Atoi(r.PathValue("id"))

	if err != nil || notificationID <= 0 {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid notification",
		})
		return
	}

	if err := notifications.MarkRead(app.DB, userID, notificationID); err != nil {
		log.Println(err)
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not mark notification read",
		})
		return
	}

	count, err := notifications.GetUnreadCount(app.DB, userID)

	if err != nil {
		log.Println(err)
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not get unread count",
		})
		return
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status": true,
		"count":  count,
	})
}
