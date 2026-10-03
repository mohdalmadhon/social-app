package api

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"social/database/chats"
	"social/database/groups"
	"social/database/notifications"
	"social/database/posts"
	"social/database/preferences"
	"social/database/users"
	"social/internal/helpers"
	"social/internal/models"
)

var mentionPattern = regexp.MustCompile(`(?:^|[^\w@])@([A-Za-z0-9_.]{2,30})`)

var chatMentionPattern = regexp.MustCompile(`(?:^|[^\w@])@([A-Za-z0-9_-]{3,12})`)

func skipsSpamCheck(n models.NewNotification) bool {
	if n.GroupInviteUserID != nil || n.EventInviteUserID != nil {
		return true
	}

	return n.PostMentionUserID != nil && n.GroupID != nil && n.PostIDTag == nil
}

// notify stores a notification for n.UserID and pushes it over the websocket
// if that user is online. It never notifies the actor about their own action.
func (app *App) notify(actorID int, n models.NewNotification) bool {
	if n.UserID <= 0 || n.UserID == actorID {
		return false
	}

	allowed, err := preferences.ShouldNotify(
		app.DB,
		n.UserID,
		actorID,
		helpers.GetNotificationPreferenceType(n),
	)

	if err != nil {
		log.Println("failed to check notification preference:", err)
	} else if !allowed {
		return false
	}

	if !skipsSpamCheck(n) {
		spam, err := notifications.IsSpam(app.DB, actorID, n)

		if err != nil {
			log.Println("failed to check notification spam:", err)
		} else if spam {
			return false
		}
	}

	id, err := notifications.InsertNotification(app.DB, n)

	if err != nil {
		log.Println("failed to insert notification:", err)
		return false
	}

	unread, err := notifications.GetUnreadCount(app.DB, n.UserID)

	if err != nil {
		log.Println("failed to count unread notifications:", err)
		unread = -1
	}

	app.sendNotification(actorID, n.UserID, n, id, unread)

	return true
}

func (app *App) actorName(userID int) string {
	user, err := users.GetUserSimpleData(app.DB, userID)

	if err != nil {
		return "Someone"
	}

	name := strings.TrimSpace(user.FirstName + " " + user.LastName)

	if name == "" {
		name = user.UserName
	}

	if name == "" {
		return "Someone"
	}

	return name
}

// mentionedUserIDs resolves @username mentions in content to user ids,
// skipping anyone already present in skip.
func (app *App) mentionedUserIDs(content string, skip map[int]bool) []int {
	var ids []int

	seen := map[int]bool{}

	for _, match := range mentionPattern.FindAllStringSubmatch(content, -1) {
		username := strings.TrimRight(match[1], ".")

		if len(username) < 2 {
			continue
		}

		id := users.GetUserID(app.DB, username)

		if id <= 0 || skip[id] || seen[id] {
			continue
		}

		seen[id] = true
		ids = append(ids, id)
	}

	return ids
}

// notifyComment covers: comment on a post, reply to a comment and
// @mentions inside the comment. Every one of them points at the post so the
// notification can show its image and open the post dialog.
func (app *App) notifyComment(actorID int, comment models.Comment) {
	name := app.actorName(actorID)
	actor := actorID
	postID := comment.PostID
	commentID := comment.ID

	notified := map[int]bool{actorID: true}

	if comment.RepltTo != nil {
		ownerID, _, err := posts.GetCommentOwner(app.DB, *comment.RepltTo)

		if err != nil {
			log.Println("could not find comment owner:", err)
		} else if !notified[ownerID] {
			notified[ownerID] = true

			app.notify(actorID, models.NewNotification{
				UserID:             ownerID,
				Message:            fmt.Sprintf("%s replied to your comment", name),
				PostIDTag:          &postID,
				CommentIDTag:       &commentID,
				CommentReplyUserID: &actor,
			})
		}
	} else {
		ownerID, err := posts.GetPostOwnerID(app.DB, postID)

		if err != nil {
			log.Println("could not find post owner:", err)
		} else if !notified[ownerID] {
			notified[ownerID] = true

			app.notify(actorID, models.NewNotification{
				UserID:             ownerID,
				Message:            fmt.Sprintf("%s commented on your post", name),
				PostIDTag:          &postID,
				CommentIDTag:       &commentID,
				CommentReplyUserID: &actor,
			})
		}
	}

	for _, id := range app.mentionedUserIDs(comment.Content, notified) {
		notified[id] = true

		app.notify(actorID, models.NewNotification{
			UserID:               id,
			Message:              fmt.Sprintf("%s mentioned you in a comment", name),
			PostIDTag:            &postID,
			CommentIDTag:         &commentID,
			CommentMentionUserID: &actor,
		})
	}
}

// notifyCommentLike tells a comment's author that somebody liked it. It only
// fires when the vote that now exists is a like (so un-liking stays silent).
func (app *App) notifyCommentLike(actorID, commentID int) {
	vote, err := posts.GetUserCommentVote(app.DB, commentID, actorID)

	if err != nil || vote != 1 {
		return
	}

	ownerID, postID, err := posts.GetCommentOwner(app.DB, commentID)

	if err != nil {
		log.Println("could not find comment owner:", err)
		return
	}

	actor := actorID

	app.notify(actorID, models.NewNotification{
		UserID:            ownerID,
		Message:           fmt.Sprintf("%s liked your comment", app.actorName(actorID)),
		PostIDTag:         &postID,
		CommentIDTag:      &commentID,
		CommentLikeUserID: &actor,
	})
}

// notifyEventInvite tells every accepted group member about a new event.
func (app *App) notifyEventInvite(actorID int, event models.GroupEvent) {
	memberIDs, err := groups.GetActiveMemberIDs(app.DB, event.GroupID)

	if err != nil {
		log.Println("could not get group members:", err)
		return
	}

	name := app.actorName(actorID)
	actor := actorID
	eventID := event.ID
	groupID := event.GroupID

	for _, id := range memberIDs {
		app.notify(actorID, models.NewNotification{
			UserID:            id,
			Message:           fmt.Sprintf("%s invited you to the event \"%s\"", name, event.Title),
			EventInviteUserID: &actor,
			EventID:           &eventID,
			GroupID:           &groupID,
		})
	}
}

// notifyChatMentions tells every group member @mentioned in a group chat
// message that they were mentioned.
func (app *App) notifyChatMentions(actorID, groupID int, content string) {
	matches := chatMentionPattern.FindAllStringSubmatch(content, -1)

	if len(matches) == 0 {
		return
	}

	isPrivate, groupName, err := chats.GetChatMeta(app.DB, groupID)

	if err != nil {
		log.Println("could not get chat meta:", err)
		return
	}

	if isPrivate {
		return
	}

	name := app.actorName(actorID)
	actor := actorID
	gid := groupID

	message := fmt.Sprintf("%s mentioned you in the group chat", name)

	if groupName != "" {
		message = fmt.Sprintf("%s mentioned you in the group chat \"%s\"", name, groupName)
	}

	seen := map[int]bool{actorID: true}

	for _, match := range matches {
		id := users.GetUserID(app.DB, strings.ToLower(match[1]))

		if id <= 0 || seen[id] {
			continue
		}

		seen[id] = true

		member, err := chats.UserInGroup(app.DB, id, groupID)

		if err != nil || !member {
			continue
		}

		app.notify(actorID, models.NewNotification{
			UserID:            id,
			Message:           message,
			PostMentionUserID: &actor,
			GroupID:           &gid,
		})
	}
}

// notifyEventResponse tells the event creator that somebody answered.
func (app *App) notifyEventResponse(actorID int, event models.GroupEvent) {
	if event.UserResponse == nil {
		return
	}

	verb := "can't go to"

	if *event.UserResponse == 1 {
		verb = "is going to"
	}

	actor := actorID

	app.notify(actorID, models.NewNotification{
		UserID:              event.Creator.ID,
		Message:             fmt.Sprintf("%s %s your event \"%s\"", app.actorName(actorID), verb, event.Title),
		EventResponseUserID: &actor,
	})
}
