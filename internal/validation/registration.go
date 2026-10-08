package validation

import (
	"errors"
	"regexp"
	"social/internal/models"
	"strings"
	"time"
)

var (
	nameRegex      = regexp.MustCompile(`^[A-Z][a-zA-Z]*$`)
	usernameRegex  = regexp.MustCompile(`^[a-z0-9_-]+$`)
	emailRegex     = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	numberRegex    = regexp.MustCompile(`[0-9]`)
	specialRegex   = regexp.MustCompile(`[!@#$%*&^?]`)
	alphaRegex     = regexp.MustCompile(`[a-zA-Z]`)
	forbiddenRegex = regexp.MustCompile(`[^a-zA-Z0-9!@#$%*&^?]`)
)

func ValidateRegisterData(userData *models.UserRegistration) error {
	NormalizeRegisterData(userData)

	if err := validateNames(&userData.FirstName); err != nil {
		return err
	}

	if err := validateNames(&userData.LastName); err != nil {
		return err
	}

	if err := validateUserName(&userData.UserName); err != nil {
		return err
	}

	if err := validateEmail(&userData.Email); err != nil {
		return err
	}

	if err := validateDOB(&userData.DOB); err != nil {
		return err
	}

	if err := validatePassword(userData.Password); err != nil {
		return err
	}

	if err := validateAbout(&userData.About); err != nil {
		return err
	}

	return nil
}

func NormalizeRegisterData(userData *models.UserRegistration) {
	userData.FirstName = strings.TrimSpace(userData.FirstName)
	userData.LastName = strings.TrimSpace(userData.LastName)
	userData.UserName = strings.TrimSpace(userData.UserName)
	userData.Email = strings.TrimSpace(userData.Email)
	userData.Password = strings.Trim(userData.Password, " ")
	userData.About = CleanText(userData.About)

	if len(userData.FirstName) >= 2 {
		userData.FirstName = strings.ToUpper(userData.FirstName[:1]) + userData.FirstName[1:]
	}

	if len(userData.LastName) >= 2 {
		userData.LastName = strings.ToUpper(userData.LastName[:1]) + userData.LastName[1:]
	}

	if userData.UserName != "" {
		userData.UserName = strings.ToLower(userData.UserName)
	}

	if userData.Email != "" {
		userData.Email = strings.ToLower(userData.Email)
	}
}

func validateNames(name *string) error {
	if RuneLen(*name) < MinName || RuneLen(*name) > MaxName {
		return errors.New("first/last name must be between 2 and 15 characters")
	}

	if !nameRegex.MatchString(*name) {
		return errors.New("name format is incorrect")
	}

	return nil
}

func validateUserName(username *string) error {
	if len(*username) == 0 {
		return nil
	}

	if RuneLen(*username) < MinUsername || RuneLen(*username) > MaxUsername {
		return errors.New("username length must be between 3 and 12 characters")
	}

	if !usernameRegex.MatchString(*username) {
		return errors.New("invalid username format")
	}

	return nil
}

func validateEmail(email *string) error {
	if RuneLen(*email) < MinEmail || RuneLen(*email) > MaxEmail {
		return errors.New("email length must be between 5 and 75 characters")
	}

	if !emailRegex.MatchString(*email) {
		return errors.New("invalid email")
	}

	return nil
}

func validateDOB(dob *time.Time) error {
	now := time.Now()

	if dob.After(now) {
		return errors.New("date of birth is invalid")
	}

	tooOld := now.AddDate(-120, 0, 0)

	if dob.Before(tooOld) {
		return errors.New("date of birth is invalid")
	}

	tooYoung := now.AddDate(-12, 0, 0)

	if dob.After(tooYoung) {
		return errors.New("date of birth is invalid")
	}

	return nil
}

func validatePassword(pass string) error {
	if RuneLen(pass) < MinPassword {
		return errors.New("invalid password: password must be at least 8 characters long")
	}

	if RuneLen(pass) > MaxPassword {
		return errors.New("invalid password: password must be less than 75 characters long")
	}

	if !numberRegex.MatchString(pass) {
		return errors.New("invalid password: password must contain a number")
	}

	if !specialRegex.MatchString(pass) {
		return errors.New("invalid password: password must contain a special character: ! @ # $ % * & ^ ?")
	}

	if !alphaRegex.MatchString(pass) {
		return errors.New("invalid password: password must contain alphabetic characters")
	}

	if forbiddenRegex.MatchString(pass) {
		return errors.New("invalid password: password contains forbidden characters")
	}

	return nil
}

func validateAbout(about *string) error {
	if len(*about) == 0 {
		return nil
	}

	return ValidateMultiline("about", *about, 0, MaxAbout, 0)
}

func ValidateLogin(identifier string, password string) (string, error) {
	identifier = strings.TrimSpace(identifier)

	if err := ValidateSingleLine("identifier", identifier, 3, MaxIdentifier); err != nil {
		return "", err
	}

	if err := ValidateSingleLine("password", password, MinPassword, MaxPassword); err != nil {
		return "", err
	}

	return identifier, nil
}

func ValidateAvailabilityValue(kind string, value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))

	switch kind {
	case "name":
		if err := ValidateSingleLine("username", value, MinUsername, MaxUsername); err != nil {
			return "", err
		}
	case "email":
		if err := ValidateSingleLine("email", value, MinEmail, MaxEmail); err != nil {
			return "", err
		}
	default:
		return "", errors.New("invalid check type")
	}

	return value, nil
}

func ValidateVerificationCode(code string) error {
	code = strings.TrimSpace(code)

	if RuneLen(code) != MaxCode {
		return errors.New("invalid code")
	}

	for _, r := range code {
		if r < '0' || r > '9' {
			return errors.New("invalid code")
		}
	}

	return nil
}

func ValidateAboutFields(about *models.UserAbout) error {
	fields := []struct {
		label string
		value *string
	}{
		{"bio", &about.Bio},
		{"work", &about.Work},
		{"education", &about.Education},
		{"travel", &about.Travel},
		{"interests", &about.Intrests},
		{"hobbies", &about.Hobbies},
		{"website", &about.Website},
		{"linkedin", &about.Linkedin},
		{"instagram", &about.Instgram},
		{"twitter", &about.Twitter},
	}

	for _, field := range fields {
		if IsLinkField(field.label) {
			link, err := NormalizeLink(field.label, *field.value)

			if err != nil {
				return err
			}

			*field.value = link
			continue
		}

		*field.value = CleanText(*field.value)

		max := MaxAboutField

		if field.label == "bio" {
			max = MaxAbout
		}

		if err := ValidateMultiline(field.label, *field.value, 0, max, 0); err != nil {
			return err
		}
	}

	return nil
}
