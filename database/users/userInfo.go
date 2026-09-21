package users

import (
	"database/sql"
	"os"

	"social/internal/models"
)

func GetUserID(db *sql.DB, identifier string) int {
	var id int

	err := db.QueryRow(`
		SELECT id
		FROM user
		WHERE username = ? OR email = ?
	`, identifier, identifier).Scan(&id)

	if err != nil {
		return -1
	}

	return id
}

func GetUserData(db *sql.DB, userID int) (models.UserData, error) {
	var userData models.UserData

	var firstName sql.NullString
	var lastName sql.NullString
	var email sql.NullString
	var username sql.NullString
	var dob sql.NullTime

	err := db.QueryRow(`
		SELECT first_name, last_name, email, username, dob
		FROM user
		WHERE id = ?
	`, userID).Scan(
		&firstName,
		&lastName,
		&email,
		&username,
		&dob,
	)

	if err != nil {
		return userData, err
	}

	if firstName.Valid {
		userData.UserInfo.FirstName = firstName.String
	} else {
		userData.UserInfo.FirstName = ""
	}

	if lastName.Valid {
		userData.UserInfo.LastName = lastName.String
	} else {
		userData.UserInfo.LastName = ""
	}

	if email.Valid {
		userData.UserInfo.Email = email.String
	} else {
		userData.UserInfo.Email = ""
	}

	if username.Valid {
		userData.UserInfo.UserName = username.String
	} else {
		userData.UserInfo.UserName = ""
	}

	if dob.Valid {
		userData.UserInfo.DOB = dob.Time
	}

	var followers sql.NullInt64
	var following sql.NullInt64
	var posts sql.NullInt64
	var avatar sql.NullString
	var about sql.NullString
	var isPrivate sql.NullInt64

	if err := db.QueryRow(`
		SELECT num_of_followers, num_of_following, num_of_posts, avatar_path, about, is_private
		FROM profile
		WHERE user_id = ?
	`, userID).Scan(
		&followers,
		&following,
		&posts,
		&avatar,
		&about,
		&isPrivate,
	); err != nil {
		return userData, err
	}

	if followers.Valid {
		userData.NumOfFollowers = int(followers.Int64)
	} else {
		userData.NumOfFollowers = 0
	}

	if following.Valid {
		userData.NumOfFollowing = int(following.Int64)
	} else {
		userData.NumOfFollowing = 0
	}

	if posts.Valid {
		userData.NumOfPosts = int(posts.Int64)
	} else {
		userData.NumOfPosts = 0
	}

	if avatar.Valid {
		userData.UserInfo.Avatar = avatar.String
	} else {
		userData.UserInfo.Avatar = ""
	}

	if about.Valid {
		userData.About.Bio = about.String
	} else {
		userData.About.Bio = ""
	}

	if isPrivate.Valid {
		userData.IsPrivate = int(isPrivate.Int64)
	} else {
		userData.IsPrivate = 0
	}

	return userData, nil
}

func UpdateUserInfo(db *sql.DB, userID int, userData *models.UserRegistration) error {
	_, err := db.Exec(`
		UPDATE user
		SET first_name = ?, last_name = ?, email = ?, username = ?
		WHERE id = ?
	`, userData.FirstName, userData.LastName, userData.Email, userData.UserName, userID)

	if err != nil {
		return err
	}

	_, err = db.Exec(`
		UPDATE profile
		SET about = ?, is_private = ?
		WHERE user_id = ?
	`, userData.About, userData.IsPrivate, userID)

	return err
}

func UpdateUserAvatar(db *sql.DB, userID int, avatar_path string) error {
	if err := DeleteOldAvatar(db, userID); err != nil {
		return err
	}

	_, err := db.Exec(`
		UPDATE profile
		SET avatar_path = ?
		WHERE user_id = ?
	`, avatar_path, userID)

	return err
}

func DeleteOldAvatar(db *sql.DB, userID int) error {
	var avatarPath string

	err := db.QueryRow(`
		SELECT avatar_path
		FROM profile
		WHERE user_id = ?
	`, userID).Scan(&avatarPath)

	if err != nil {
		return err
	}

	if avatarPath == "" || avatarPath == "avatars/default.png" {
		return nil
	}

	err = os.Remove("./uploads/" + avatarPath)

	if os.IsNotExist(err) {
		return nil
	}

	return err
}


// DeleteUser removes a user's account permanently. Related rows
// (profile, posts, comments, messages, notifications, etc.) are removed
// automatically through ON DELETE CASCADE foreign keys.
func DeleteUser(db *sql.DB, userID int) error {
	if err := DeleteOldAvatar(db, userID); err != nil {
		// missing/default avatar should not block account deletion
		if err != sql.ErrNoRows {
			return err
		}
	}

	_, err := db.Exec(`
		DELETE FROM user
		WHERE id = ?
	`, userID)

	return err
}

func UserExists(db *sql.DB, userID int) error {
	var id int

	return db.QueryRow(
		`SELECT id FROM user WHERE id = ?`,
		userID,
	).Scan(&id)
}

func GetUserSimpleData(db *sql.DB, userID int) (models.UserRegistration, error) {
	var user models.UserRegistration

	var firstName sql.NullString
	var lastName sql.NullString

	err := db.QueryRow(`
		SELECT id, first_name, last_name
		FROM user
		WHERE id = ?
	`, userID).Scan(
		&user.ID,
		&firstName,
		&lastName,
	)

	if err != nil {
		return models.UserRegistration{}, err
	}

	if firstName.Valid {
		user.FirstName = firstName.String
	} else {
		user.FirstName = ""
	}

	if lastName.Valid {
		user.LastName = lastName.String
	} else {
		user.LastName = ""
	}

	var avatar sql.NullString

	err = db.QueryRow(`
		SELECT avatar_path
		FROM profile
		WHERE user_id = ?
	`, userID).Scan(&avatar)

	if err != nil {
		return user, err
	}

	if avatar.Valid {
		user.Avatar = avatar.String
	} else {
		user.Avatar = ""
	}

	return user, nil
}
