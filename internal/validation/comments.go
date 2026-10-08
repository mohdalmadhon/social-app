package validation

import (
	"errors"
	"social/internal/models"
)

func ValidateComment(comment *models.Comment, hasImage bool) error {
	comment.Content = CleanText(comment.Content)

	if comment.PostID <= 0 {
		return errors.New("invalid post")
	}

	if comment.RepltTo != nil && *comment.RepltTo <= 0 {
		return errors.New("invalid reply")
	}

	if len(comment.Content) == 0 && !hasImage {
		return errors.New("comment cannot be empty")
	}

	return ValidateMultiline("comment", comment.Content, 0, MaxComment, MaxCommentLines)
}
