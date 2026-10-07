package models

import "time"

type Group struct {
	ID          int
	UserID      int
	Name        string `json:"name"`
	Count       int
	UnreadCount int `json:"unreadCount"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Avatar      string `json:"avatar"`
	Users       []UserRegistration
	CreatedAt   time.Time
}

type NewGroup struct {
	GroupID int    `json:"groupId"`
	UserID  int    `json:"userId"`
	Name    string `json:"name"`
	Users   []int  `json:"users"`
}
