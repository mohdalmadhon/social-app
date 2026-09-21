package models

type SearchUser struct {
	ID        int    `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Username  string `json:"username"`
	Avatar    string `json:"avatar"`
	IsPrivate bool   `json:"isPrivate"`
	// FollowStatus is the searching user's follow state towards this user:
	// -1 not following, 0 request pending, 1 following.
	FollowStatus int  `json:"followStatus"`
	IsFriend     bool `json:"isFriend"`
	IsMe         bool `json:"isMe"`
}

type SearchGroup struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Avatar       string `json:"avatar"`
	MembersCount int    `json:"membersCount"`
	IsMember     bool   `json:"isMember"`
}
