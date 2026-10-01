package preferences

import (
	"database/sql"
	"log"
)

func ChangeChatPreferences(db *sql.DB, value string, userID int) error {
	log.Println(value)
	_, err := db.Exec(`update user_preferences set chat = ? where user_id = ?`,value, userID)
	return err
}
