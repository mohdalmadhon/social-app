package models

import "time"

type ChatList struct {
	ID      int
	Members []UserRegistration
	Name    string
	Avatar  string
}

type PrivateChat struct {
	UserID    int
	GroupID   int
	Avatar    string
	FirstName string
	LastName  string
}

type Message struct {
	ID        int
	Sender    UserRegistration
	GroupID   int
	Content   string `json:"content"`
	CreatedAt time.Time
}

