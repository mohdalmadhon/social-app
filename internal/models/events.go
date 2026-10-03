package models

type GroupEvent struct {
	ID           int              `json:"id"`
	GroupID      int              `json:"groupId"`
	Title        string           `json:"title"`
	Description  string           `json:"description"`
	EventTime    string           `json:"eventTime"`
	CreatedAt    string           `json:"createdAt"`
	Creator      UserRegistration `json:"creator"`
	GoingCount   int              `json:"goingCount"`
	NotGoingCount int             `json:"notGoingCount"`
	UserResponse *int             `json:"userResponse"`
}

type NewGroupEvent struct {
	GroupID     int    `json:"groupId"`
	Title       string `json:"title"`
	Description string `json:"description"`
	EventTime   string `json:"eventTime"`
	TzOffset    *int   `json:"tzOffset"`
}

type GroupEventResponse struct {
	EventID  int `json:"eventId"`
	Response int `json:"response"`
}

type GroupEventVoter struct {
	User      UserRegistration `json:"user"`
	Response  int              `json:"response"`
	CreatedAt string           `json:"createdAt"`
}
