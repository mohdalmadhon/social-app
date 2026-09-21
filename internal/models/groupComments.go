package models

type GroupComment struct {
	ID          int              `json:"id"`
	Content     string           `json:"content"`
	User        UserRegistration `json:"user"`
	GroupPostID int              `json:"postId"`
	ReplyTo     *int             `json:"replyTo"`
	Votes       int              `json:"votes"`
	CreatedAt   string           `json:"createdAt"`
	Replies     int              `json:"replies"`
}
