package models

import "time"

type ChatList struct {
	ID      int
	Members []UserRegistration
	Name    string
	Avatar  string
}

type PrivateChat struct {
	UserID     int
	GroupID    int
	Avatar     string
	FirstName  string
	LastName   string
	CanMessage bool
}

type Message struct {
	ID        int
	Sender    UserRegistration
	GroupID   int
	Content   string `json:"content"`
	CreatedAt time.Time
}

type SentPost struct {
	PostID     int `json:"postID"`
	UserID     int `json:"userID"`
	RecieverID int `json:"recieverID"`
}
