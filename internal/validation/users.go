package validation

import (
	"errors"
	"social/internal/models"
)

func ValidateUpdateInfo(userData *models.UserRegistration) error {
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

	if len(userData.Password) != 0 {
		if err := validatePassword(userData.Password); err != nil {
			return err
		}
	}

	if err := validateAbout(&userData.About); err != nil {
		return err
	}

	if userData.IsPrivate != 1 && userData.IsPrivate != 0 {
		return errors.New("invalid is private value")
	}

	return nil
}
