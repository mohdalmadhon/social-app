package preferences

import (
	"database/sql"
	"fmt"
	"slices"

	"social/database/users"
)

var NotificationTypes = []string{
	"follow",
	"post_reaction",
	"comment",
	"comment_like",
	"mention",
	"event",
	"message",
}

var NotificationModes = []string{"any", "friends", "none"}

func IsNotificationType(value string) bool {
	return slices.Contains(NotificationTypes, value)
}

func IsNotificationMode(value string) bool {
	return slices.Contains(NotificationModes, value)
}

func GetNotificationPreferences(db *sql.DB, userID int) (map[string]string, error) {
	prefs := make(map[string]string, len(NotificationTypes))

	for _, notificationType := range NotificationTypes {
		prefs[notificationType] = "any"
	}

	rows, err := db.Query(`
		SELECT type, mode
		FROM notification_preferences
		WHERE user_id = ?
	`, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var notificationType, mode string

		if err := rows.Scan(&notificationType, &mode); err != nil {
			return nil, err
		}

		prefs[notificationType] = mode
	}

	return prefs, rows.Err()
}

func SetNotificationPreference(db *sql.DB, userID int, notificationType, mode string) error {
	if !IsNotificationType(notificationType) || !IsNotificationMode(mode) {
		return fmt.Errorf("invalid notification preference")
	}

	_, err := db.Exec(`
		INSERT INTO notification_preferences (user_id, type, mode)
		VALUES (?, ?, ?)
		ON CONFLICT(user_id, type) DO UPDATE SET mode = excluded.mode
	`, userID, notificationType, mode)

	return err
}

func ShouldNotify(db *sql.DB, recipientID, actorID int, notificationType string) (bool, error) {
	if notificationType == "" {
		return true, nil
	}

	var mode string

	err := db.QueryRow(`
		SELECT mode
		FROM notification_preferences
		WHERE user_id = ? AND type = ?
	`, recipientID, notificationType).Scan(&mode)

	if err == sql.ErrNoRows {
		return true, nil
	}

	if err != nil {
		return true, err
	}

	switch mode {
	case "none":
		return false, nil

	case "friends":
		return users.IsFriend(db, recipientID, actorID)
	}

	return true, nil
}
