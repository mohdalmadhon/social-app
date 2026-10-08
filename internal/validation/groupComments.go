package validation

import (
	"errors"
	"social/internal/models"
)

func ValidateGroupComment(comment *models.GroupComment, hasImage bool) error {
	comment.Content = CleanText(comment.Content)

	if comment.GroupPostID <= 0 {
		return errors.New("invalid group post")
	}

	if comment.ReplyTo != nil && *comment.ReplyTo <= 0 {
		return errors.New("invalid reply")
	}

	if len(comment.Content) == 0 && !hasImage {
		return errors.New("comment cannot be empty")
	}

	return ValidateMultiline("comment", comment.Content, 0, MaxComment, MaxCommentLines)
}
