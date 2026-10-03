package validation

import (
	"errors"
	"strings"
	"time"

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

	eventTime, err := parseEventTime(event.EventTime, event.TzOffset)

	if err != nil {
		return errors.New("invalid event day and time")
	}

	if !eventTime.After(time.Now()) {
		return errors.New("event cannot be in the past")
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

func parseEventTime(value string, tzOffset *int) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t, nil
	}

	layouts := []string{"2006-01-02T15:04", "2006-01-02T15:04:05", "2006-01-02 15:04"}

	for _, layout := range layouts {
		if tzOffset == nil {
			if t, err := time.ParseInLocation(layout, value, time.Local); err == nil {
				return t, nil
			}

			continue
		}

		if t, err := time.Parse(layout, value); err == nil {
			return t.Add(time.Duration(*tzOffset) * time.Minute), nil
		}
	}

	return time.Time{}, errors.New("invalid time")
}
