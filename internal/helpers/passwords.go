package helpers

import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashedPassword), nil
}

func AuthonticateUser(pass string, hashedPassword string) bool {
	if err := (bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(pass))); err != nil {
		return false
	}
	return true
}
