package validation

import (
	"encoding/binary"
	"errors"
	"io"
	"mime/multipart"
	"path/filepath"
	"social/internal/models"
	"strconv"
	"strings"
)

type ValidationResult struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func ValidatePost(data *models.RegsiterPost, image *multipart.FileHeader) ValidationResult {
	data.Content = CleanText(data.Content)
	data.Location = strings.TrimSpace(data.Location)

	if data.Content == "" {
		return ValidationResult{
			Field:   "content",
			Message: "content cannot be empty",
		}
	}

	if err := ValidateMultiline("content", data.Content, 1, MaxPostContent, 0); err != nil {
		return ValidationResult{
			Field:   "content",
			Message: err.Error(),
		}
	}

	if len(data.PeopleTagged) > MaxPostTags {
		return ValidationResult{
			Field:   "tags",
			Message: "too many tagged people",
		}
	}

	if RuneLen(data.Location) > MaxPostLocation {
		return ValidationResult{
			Field:   "location",
			Message: "location is too long",
		}
	}

	if data.AllowComments != 0 && data.AllowComments != 1 {
		return ValidationResult{
			Field:   "allowComments",
			Message: "invalid allow comments code",
		}
	}

	if data.GroupID < -1 {
		return ValidationResult{
			Field:   "groupID",
			Message: "invalid group code",
		}
	}

	if data.Location != "" {
		locationParts := strings.Split(data.Location, ":")

		if len(locationParts) != 3 ||
			strings.TrimSpace(locationParts[0]) == "" ||
			strings.TrimSpace(locationParts[1]) == "" ||
			strings.TrimSpace(locationParts[2]) == "" {

			return ValidationResult{
				Field:   "location",
				Message: "invalid location",
			}
		}

		if _, err := strconv.ParseFloat(strings.TrimSpace(locationParts[1]), 64); err != nil {
			return ValidationResult{
				Field:   "location",
				Message: "invalid location",
			}
		}

		if _, err := strconv.ParseFloat(strings.TrimSpace(locationParts[2]), 64); err != nil {
			return ValidationResult{
				Field:   "location",
				Message: "invalid location",
			}
		}
	}

	for _, id := range data.PeopleTagged {
		if id <= 0 {
			return ValidationResult{
				Field:   "tags",
				Message: "invalid tagged people",
			}
		}
	}

	if image != nil {
		allowedExtensions := map[string]bool{
			".png":  true,
			".jpg":  true,
			".jpeg": true,
			".gif":  true,
			".mp4":  true,
		}

		extension := strings.ToLower(filepath.Ext(image.Filename))

		if !allowedExtensions[extension] {
			return ValidationResult{
				Field:   "image",
				Message: "image must be png, jpg, jpeg, gif or mp4",
			}
		}

		if image.Size <= 0 {
			return ValidationResult{
				Field:   "image",
				Message: "invalid image",
			}
		}

		if image.Size > MaxPostMediaSize {
			return ValidationResult{
				Field:   "image",
				Message: "file is too large",
			}
		}

		if extension == ".mp4" {
			file, err := image.Open()
			if err != nil {
				return ValidationResult{
					Field:   "image",
					Message: "could not read video",
				}
			}
			defer file.Close()

			duration, err := getMP4Duration(file)
			if err != nil {
				return ValidationResult{
					Field:   "image",
					Message: "invalid mp4 video",
				}
			}

			if duration > 60 {
				return ValidationResult{
					Field:   "image",
					Message: "video cannot be longer than 60 seconds",
				}
			}
		}
	}

	return ValidationResult{}
}

func getMP4Duration(r io.ReaderAt) (float64, error) {
	var offset int64

	for {
		header := make([]byte, 8)

		_, err := r.ReadAt(header, offset)
		if err != nil {
			return 0, err
		}

		size := binary.BigEndian.Uint32(header[0:4])
		boxType := string(header[4:8])

		headerSize := int64(8)
		boxSize := int64(size)

		if size == 1 {
			extendedSize := make([]byte, 8)

			_, err := r.ReadAt(extendedSize, offset+8)
			if err != nil {
				return 0, err
			}

			boxSize = int64(binary.BigEndian.Uint64(extendedSize))
			headerSize = 16
		}

		if size == 0 {
			return 0, errors.New("invalid mp4 box size")
		}

		if boxSize < headerSize {
			return 0, errors.New("invalid mp4 box")
		}

		switch boxType {
		case "moov":
			duration, err := findMP4Duration(
				r,
				offset+headerSize,
				boxSize-headerSize,
			)

			if err == nil {
				return duration, nil
			}
		}

		offset += boxSize
	}
}

func findMP4Duration(
	r io.ReaderAt,
	offset int64,
	size int64,
) (float64, error) {
	end := offset + size

	for offset < end {
		header := make([]byte, 8)

		_, err := r.ReadAt(header, offset)
		if err != nil {
			return 0, err
		}

		boxSize := int64(binary.BigEndian.Uint32(header[0:4]))
		boxType := string(header[4:8])
		headerSize := int64(8)

		if boxSize == 1 {
			extendedSize := make([]byte, 8)

			_, err := r.ReadAt(extendedSize, offset+8)
			if err != nil {
				return 0, err
			}

			boxSize = int64(binary.BigEndian.Uint64(extendedSize))
			headerSize = 16
		}

		if boxSize < headerSize {
			return 0, errors.New("invalid mp4 box")
		}

		dataOffset := offset + headerSize
		dataSize := boxSize - headerSize

		switch boxType {
		case "mvhd":
			return readMVHD(r, dataOffset, dataSize)

		case "trak", "mdia":
			duration, err := findMP4Duration(
				r,
				dataOffset,
				dataSize,
			)

			if err == nil {
				return duration, nil
			}
		}

		offset += boxSize
	}

	return 0, errors.New("mp4 duration not found")
}

func readMVHD(
	r io.ReaderAt,
	offset int64,
	size int64,
) (float64, error) {
	if size < 20 {
		return 0, errors.New("invalid mvhd")
	}

	version := make([]byte, 1)

	_, err := r.ReadAt(version, offset)
	if err != nil {
		return 0, err
	}

	if version[0] == 0 {
		data := make([]byte, 12)

		_, err := r.ReadAt(data, offset+4)
		if err != nil {
			return 0, err
		}

		timescale := binary.BigEndian.Uint32(data[4:8])
		duration := binary.BigEndian.Uint32(data[8:12])

		if timescale == 0 {
			return 0, errors.New("invalid timescale")
		}

		return float64(duration) / float64(timescale), nil
	}

	if version[0] == 1 {
		data := make([]byte, 20)

		_, err := r.ReadAt(data, offset+4)
		if err != nil {
			return 0, err
		}

		timescale := binary.BigEndian.Uint32(data[16:20])

		durationData := make([]byte, 8)

		_, err = r.ReadAt(durationData, offset+20)
		if err != nil {
			return 0, err
		}

		duration := binary.BigEndian.Uint64(durationData)

		if timescale == 0 {
			return 0, errors.New("invalid timescale")
		}

		return float64(duration) / float64(timescale), nil
	}

	return 0, errors.New("unsupported mvhd version")
}
