package validation

import (
	"errors"
	"social/internal/models"
)

func ValidateGroupComment(comment models.GroupComment, hasImage bool) error {
	if comment.GroupPostID <= 0 {
		return errors.New("invalid group post")
	}

	if len(comment.Content) <= 0 && !hasImage {
		return errors.New("comment cannot be empty")
	}

	if len(comment.Content) > 200 {
		return errors.New("comment cannot be more than 200 character")
	}

	return nil
}
