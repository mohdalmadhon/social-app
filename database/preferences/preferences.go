package preferences

import (
	"database/sql"
	"fmt"
)

type Preferences struct {
	Chat           string `json:"chat"`
	Email          string `json:"email"`
	DOB            string `json:"dob"`
	AdditionalInfo string `json:"additionalInfo"`
	GroupInvite    string `json:"groupInvite"`

	AllowPreviousSenders bool `json:"allowPreviousSenders"`
}

var columns = map[string]string{
	"chat":       "chat",
	"email":      "show_email",
	"dob":        "show_dob",
	"additional": "additional_info",
	"groupinvite": "group_invite",
	"previoussenders": "allow_previous_senders",
}

func ChangePreference(db *sql.DB, prefType, value string, userID int) error {
	column, ok := columns[prefType]
	if !ok {
		return fmt.Errorf("unknown preference type")
	}

	_, err := db.Exec(
		`UPDATE user_preferences SET `+column+` = ? WHERE user_id = ?`,
		value, userID,
	)
	return err
}

func ChangeChatPreferences(db *sql.DB, value string, userID int) error {
	return ChangePreference(db, "chat", value, userID)
}

func GetPreferences(db *sql.DB, userID int) (Preferences, error) {
	var p Preferences
	var allowPreviousSenders int

	err := db.QueryRow(`
		SELECT chat, show_email, show_dob, additional_info, group_invite, allow_previous_senders
		FROM user_preferences
		WHERE user_id = ?
	`, userID).Scan(&p.Chat, &p.Email, &p.DOB, &p.AdditionalInfo, &p.GroupInvite, &allowPreviousSenders)

	p.AllowPreviousSenders = allowPreviousSenders == 1

	return p, err
}
