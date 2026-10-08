package api

import (
	"encoding/json"
	"errors"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"social/internal/helpers"
)

var errInvalidComment = errors.New("invalid comment")
var errInvalidCommentForm = errors.New("invalid form data or image is too large")

type commentInput struct {
	PostID  int
	Content string
	ReplyTo *int
	form    *multipart.Form
	file    multipart.File
	header  *multipart.FileHeader
}

func readCommentInput(w http.ResponseWriter, r *http.Request) (*commentInput, error) {
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		r.Body = http.MaxBytesReader(w, r.Body, 64<<10)

		var body struct {
			PostID  int    `json:"postId"`
			Content string `json:"content"`
			ReplyTo *int   `json:"replyTo"`
		}

		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			return nil, errInvalidComment
		}

		return &commentInput{
			PostID:  body.PostID,
			Content: body.Content,
			ReplyTo: body.ReplyTo,
		}, nil
	}

	r.Body = http.MaxBytesReader(w, r.Body, helpers.MaxChatMediaSize+(1<<20))

	if err := r.ParseMultipartForm(helpers.MaxChatMediaSize); err != nil {
		return nil, errInvalidCommentForm
	}

	input := &commentInput{
		Content: r.FormValue("content"),
		form:    r.MultipartForm,
	}

	postID, err := strconv.Atoi(r.FormValue("postId"))

	if err != nil {
		input.Close()
		return nil, errInvalidComment
	}

	input.PostID = postID

	if value := r.FormValue("replyTo"); value != "" {
		replyTo, err := strconv.Atoi(value)

		if err != nil {
			input.Close()
			return nil, errInvalidComment
		}

		input.ReplyTo = &replyTo
	}

	file, header, err := r.FormFile("image")

	if err == nil {
		input.file = file
		input.header = header
	} else if !errors.Is(err, http.ErrMissingFile) {
		input.Close()
		return nil, errInvalidComment
	}

	return input, nil
}

func (c *commentInput) HasImage() bool {
	return c.file != nil
}

func (c *commentInput) SaveImage() (string, error) {
	return helpers.SaveCommentMedia(c.file, c.header)
}

func (c *commentInput) Close() {
	if c.file != nil {
		c.file.Close()
	}

	if c.form != nil {
		c.form.RemoveAll()
	}
}

func writeCommentMediaError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	message := "could not save image"

	if errors.Is(err, helpers.ErrChatMediaTooLarge) || errors.Is(err, helpers.ErrChatMediaUnsupported) {
		status = http.StatusBadRequest
		message = err.Error()
	} else {
		log.Println(err)
	}

	helpers.WriteJson(w, status, map[string]any{
		"status":  false,
		"message": message,
	})
}
