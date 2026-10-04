package helpers

import (
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

const AVATAR_PATH = "uploads/avatars"
const POSTS_PATH = "uploads/posts"
const AVATARS_GROUPS_PATH = "uploads/groups/avatars"
const CHATS_PATH = "uploads/chats"
const COMMENTS_PATH = "uploads/comments"
const MaxChatMediaSize = 8 << 20

var ErrChatMediaTooLarge = errors.New("file is too large")
var ErrChatMediaUnsupported = errors.New("only jpg, png, gif and webp images are allowed")

var chatMediaTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

func SaveUploads(file multipart.File, header *multipart.FileHeader, Type string) (string, error) {
	var path string

	if Type == "post" {
		path = POSTS_PATH
	} else if Type == "avatar" {
		path = AVATAR_PATH
	} else if Type == "group/avatar" {
		path = AVATARS_GROUPS_PATH
	} else {
		return "", fmt.Errorf("invalid upload type")
	}

	if err := os.MkdirAll(path, 0755); err != nil {
		return "", err
	}

	extension := filepath.Ext(header.Filename)
	filename := uuid.New().String() + extension
	filePath := filepath.Join(path, filename)

	log.Println("Saving upload to:", filePath)

	dst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}

	defer dst.Close()

	_, err = io.Copy(dst, file)

	if err != nil {
		return "", err
	}

	if Type == "avatar" {
		return "avatars/" + filename, nil
	}

	return "posts/" + filename, nil
}

func SaveChatMedia(file multipart.File, header *multipart.FileHeader) (string, string, error) {
	return saveMedia(file, header, CHATS_PATH, "chats/")
}

func SaveCommentMedia(file multipart.File, header *multipart.FileHeader) (string, error) {
	path, _, err := saveMedia(file, header, COMMENTS_PATH, "comments/")
	return path, err
}

func RemoveCommentMedia(path string) {
	if !strings.HasPrefix(path, "comments/") {
		return
	}

	filePath := filepath.Join(COMMENTS_PATH, filepath.Base(path))

	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		log.Println("could not remove comment media:", err)
	}
}

func saveMedia(file multipart.File, header *multipart.FileHeader, dir, prefix string) (string, string, error) {
	if header.Size > MaxChatMediaSize {
		return "", "", ErrChatMediaTooLarge
	}

	buffer := make([]byte, 512)

	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", "", err
	}

	contentType := http.DetectContentType(buffer[:n])

	extension, ok := chatMediaTypes[contentType]
	if !ok {
		return "", "", ErrChatMediaUnsupported
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", "", err
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", "", err
	}

	filename := uuid.New().String() + extension
	filePath := filepath.Join(dir, filename)

	dst, err := os.Create(filePath)
	if err != nil {
		return "", "", err
	}

	written, err := io.Copy(dst, io.LimitReader(file, MaxChatMediaSize+1))
	closeErr := dst.Close()

	if err == nil {
		err = closeErr
	}

	if err == nil && written > MaxChatMediaSize {
		err = ErrChatMediaTooLarge
	}

	if err != nil {
		os.Remove(filePath)
		return "", "", err
	}

	kind := "image"
	if contentType == "image/gif" {
		kind = "gif"
	}

	return prefix + filename, kind, nil
}