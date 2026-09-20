package models

import "encoding/json"

type WSPayload struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type NewNotification struct {
	ID                        int
	Message                   string
	UserID                    int
	MessageUserID             *int
	PostIDTag                 *int
	CommentIDTag              *int
	CommentReplyUserID        *int
	FollowRequestUserID       *int
	FollowRequestAcceptUserID *int
	FollowUserID              *int
	PostLikeUserID            *int
	PostDislikeUserID         *int
	CommentLikeUserID         *int
	CommentMentionUserID      *int
	PostMentionUserID         *int
	GroupInviteUserID         *int
	GroupJoinUserID           *int
	GroupAcceptUserID         *int
	EventInviteUserID         *int
	EventResponseUserID       *int
}

type IncomingMessage struct {
	Offset  int    `json:"offset"`
	Private int    `json:"private"`
	UserID  int    `json:"userID"`
	GroupID int    `json:"groupID"`
	Content string `json:"content"`
}

type GroupInvite struct {
	GroupData Group `json:"groupData"`
	Users     []int `json:"users"`
}

type PostMessage struct {
	Content   string `json:"content"`
	ImagePath string `json:"imagePath"`
	PostID    int    `json:"postID"`
	GroupID   int    `json:"groupID"`
	User      User   `json:"user"`
}

type User struct {
	ID        int    `json:"ID"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Avatar    string `json:"avatar"`
}