package validation

import "strings"

func ValidateEmail(email string) error {
	email = strings.ToLower(strings.TrimSpace(email))

	return validateEmail(&email)
}
