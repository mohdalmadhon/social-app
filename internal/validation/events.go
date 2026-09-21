package validation

import (
	"errors"
	"strings"

	"social/internal/models"
)

func ValidateGroupEvent(event models.NewGroupEvent) error {
	if event.GroupID <= 0 {
		return errors.New("invalid group")
	}

	title := strings.TrimSpace(event.Title)

	if len(title) == 0 {
		return errors.New("title cannot be empty")
	}

	if len(title) > 100 {
		return errors.New("title cannot be more than 100 character")
	}

	if len(event.Description) > 1000 {
		return errors.New("description cannot be more than 1000 character")
	}

	if len(strings.TrimSpace(event.EventTime)) == 0 {
		return errors.New("event day and time are required")
	}

	return nil
}

func ValidateGroupEventResponse(response models.GroupEventResponse) error {
	if response.EventID <= 0 {
		return errors.New("invalid event")
	}

	if response.Response != 0 && response.Response != 1 {
		return errors.New("invalid response")
	}

	return nil
}
