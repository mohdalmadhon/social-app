package models

// group id == 0 (PUBLIC)
// group id == -1 (PRIVATE)
type RegsiterPost struct {
	UserID        int
	GroupID       int    `json:"groupID"`
	Content       string `json:"content"`
	Image_path    string
	AllowComments int    `json:"allowComments"`
	Location      string `json:"location"`
	PeopleTagged  []int  `json:"taggedPeople"`
}

type TaggedPerson struct {
	Id         int    `json:"id"`
	FirstName  string `json:"firstName"`
	LastName   string `json:"lastName"`
	AvatarPath string `json:"avatarPath"`
}

type Post struct {
	Public         int
	Private        int
	Id             int     `json:"id"`
	UserId         int     `json:"userId"`
	FirstName      string  `json:"firstName"`
	LastName       string  `json:"lastName"`
	Username       *string `json:"username"`
	AvatarPath     string  `json:"avatarPath"`
	Content        string  `json:"content"`
	ImagePath      *string `json:"imagePath"`
	AllowComments  bool    `json:"allowComments"`
	Location       *string `json:"location"`
	GroupId        *int    `json:"groupId"`
	GroupName      string
	CreatedAt      string         `json:"createdAt"`
	Relationship   string         `json:"relationship"`
	Visibility     string         `json:"visibility"`
	VisibilityUser string         `json:"visibilityUser"`
	TaggedPeople   []TaggedPerson `json:"taggedPeople"`
	LikeCount      int            `json:"likeCount"`
	DisLikeCount   int            `json:"disLikeCount"`
	CommentCount   int            `json:"commentCount"`
	ReactionValue  int
}

type Comment struct {
	ID        int              `json:"id"`
	Content   string           `json:"content"`
	User      UserRegistration `json:"user"`
	PostID    int              `json:"postId"`
	RepltTo   *int             `json:"replyTo"`
	Votes     int              `json:"votes"`
	CreatedAt string           `json:"createdAt"`
	Replies   int              `json:"replies"`
}

type Reaction struct {
	UserID int `json:"userID"`
	PostID int `json:"postID"`
	Value  int `json:"value"`
}
