package validation

import (
	"errors"
	"social/internal/models"
)

func ValidateComment(comment models.Comment, hasImage bool) error {
	if len(comment.Content) <= 0 && !hasImage {
		return errors.New("comment cannot be empty")
	}

	if len(comment.Content) > 200 {
		return errors.New("comment cannot be more than 200 character")
	}
	
	return nil
}