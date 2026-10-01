package users

import (
	"database/sql"
	"errors"
)

func IsAvailable(db *sql.DB, field, value string) (bool, error) {
	var query string

	switch field {
	case "email":
		query = `SELECT COUNT(1) FROM user WHERE email = ?`
	case "username":
		query = `SELECT COUNT(1) FROM user WHERE username = ?`
	default:
		return false, errors.New("invalid field")
	}

	var count int

	if err := db.QueryRow(query, value).Scan(&count); err != nil {
		return false, err
	}

	return count == 0, nil
}
