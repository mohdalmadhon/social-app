package validation

import (
	"errors"
	"mime/multipart"
	"path/filepath"
	"social/internal/models"
	"strings"
)

func ValidateGroup(g *models.Group, image *multipart.FileHeader) error {
	g.Title = strings.TrimSpace(g.Title)
	g.Description = CleanText(g.Description)

	if err := ValidateSingleLine("name", g.Title, 1, MaxGroupTitle); err != nil {
		return err
	}

	if err := ValidateMultiline("description", g.Description, 1, MaxGroupDescription, 0); err != nil {
		return err
	}

	if image != nil {
		allowedExtensions := map[string]bool{
			".png":  true,
			".jpg":  true,
			".jpeg": true,
			".gif":  true,
		}

		extension := strings.ToLower(filepath.Ext(image.Filename))

		if !allowedExtensions[extension] {
			return errors.New("image must be png, jpg or gif")
		}

		if image.Size <= 0 {
			return errors.New("invalid image")
		}

		if image.Size > MaxAvatarSize {
			return errors.New("image must be smaller than 5MB")
		}
	}

	return nil
}

func ValidateGroupName(name string) (string, error) {
	name = strings.TrimSpace(name)

	if err := ValidateSingleLine("name", name, 1, MaxGroupTitle); err != nil {
		return "", err
	}

	return name, nil
}

func ValidateNewGroup(group *models.NewGroup) error {
	name, err := ValidateGroupName(group.Name)

	if err != nil {
		return err
	}

	group.Name = name

	return ValidateIDList("group members", group.Users)
}
