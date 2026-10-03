package posts

import (
	"database/sql"
	"social/internal/models"
)

func InsertReaction(db *sql.DB, reaction models.Reaction) (error, bool) {

	var currentValue int
	err := db.QueryRow(`
		SELECT value
		FROM post_reactions
		WHERE user_id = ? AND post_id = ?
	`, reaction.UserID, reaction.PostID).Scan(&currentValue)

	if err != nil && err != sql.ErrNoRows {
		return err, false
	}

	if reaction.Value == currentValue {
		_, err := db.Exec(`
			DELETE FROM post_reactions
			WHERE post_id = ? AND user_id = ?
		`, reaction.PostID, reaction.UserID)

		return err, true
	}

	if err == sql.ErrNoRows {
		_, err := db.Exec(`
			INSERT INTO post_reactions (post_id, user_id, value)
			VALUES (?, ?, ?)
		`, reaction.PostID, reaction.UserID, reaction.Value)

		return err, false
	}

	_, err = db.Exec(`
		UPDATE post_reactions
		SET value = ?
		WHERE user_id = ? AND post_id = ?
	`, reaction.Value, reaction.UserID, reaction.PostID)

	return err, false
}
