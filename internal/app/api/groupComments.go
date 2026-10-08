package api

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"social/database/groups"
	"social/internal/helpers"
	"social/internal/models"
	"social/internal/validation"
)

func (app *App) AddGroupComment(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)

	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	input, err := readCommentInput(w, r)

	if err != nil {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": err.Error(),
		})
		return
	}

	defer input.Close()

	comment := models.GroupComment{
		Content:     input.Content,
		GroupPostID: input.PostID,
		ReplyTo:     input.ReplyTo,
	}

	comment.User.ID = userID

	if err := validation.ValidateGroupComment(&comment, input.HasImage()); err != nil {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": err.Error(),
		})
		return
	}

	allowed, err := groups.UserInPostGroup(app.DB, comment.GroupPostID, userID)

	if err != nil {
		log.Println(err)

		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "failed to insert comment",
		})
		return
	}

	if !allowed {
		helpers.WriteJson(w, http.StatusForbidden, map[string]any{
			"status":  false,
			"message": "you are not a member of this group",
		})
		return
	}

	if comment.ReplyTo != nil {
		parentPostID, err := groups.GroupCommentPostID(app.DB, *comment.ReplyTo)

		if err == sql.ErrNoRows || (err == nil && parentPostID != comment.GroupPostID) {
			helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
				"status":  false,
				"message": "invalid reply",
			})
			return
		}

		if err != nil {
			log.Println(err)

			helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
				"status":  false,
				"message": "failed to insert comment",
			})
			return
		}
	}

	if input.HasImage() {
		comment.ImagePath, err = input.SaveImage()

		if err != nil {
			writeCommentMediaError(w, err)
			return
		}
	}

	imagePath := comment.ImagePath

	comment, err = groups.InsertGroupComment(app.DB, comment)

	if err != nil {
		log.Println(err)

		if comment.ID == 0 {
			helpers.RemoveCommentMedia(imagePath)
		}

		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "failed to insert comment",
		})
		return
	}

	app.notifyGroupComment(userID, comment)

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status":  true,
		"message": "comment inserted!",
		"comment": comment,
	})
}

func (app *App) GetGroupComments(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)

	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	postID, err := strconv.Atoi(
		r.URL.Query().Get("postId"),
	)

	if err != nil || postID <= 0 {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid post id",
		})
		return
	}

	replyTo := 0

	replyValue := r.URL.Query().Get("replyTo")

	if replyValue != "" {
		replyTo, err = strconv.Atoi(replyValue)

		if err != nil {
			helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
				"status":  false,
				"message": "invalid reply id",
			})
			return
		}
	}

	allowed, err := groups.UserInPostGroup(app.DB, postID, userID)

	if err != nil {
		log.Println(err)

		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "failed to get comments",
		})
		return
	}

	if !allowed {
		helpers.WriteJson(w, http.StatusForbidden, map[string]any{
			"status":  false,
			"message": "you are not a member of this group",
		})
		return
	}

	comments, err := groups.GetGroupComments(
		app.DB,
		postID,
		replyTo,
		50,
		0,
		"latest",
	)

	if err != nil {
		log.Println(err)

		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "failed to get comments",
		})
		return
	}

	if comments == nil {
		comments = []models.GroupComment{}
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status":   true,
		"userId":   userID,
		"comments": comments,
	})
}

func (app *App) DeleteGroupComment(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)

	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	commentID, err := strconv.Atoi(
		r.URL.Query().Get("commentId"),
	)

	if err != nil || commentID <= 0 {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid comment id",
		})
		return
	}

	imagePaths, err := groups.DeleteGroupComment(
		app.DB,
		commentID,
		userID,
	)

	if err == sql.ErrNoRows {
		helpers.WriteJson(w, http.StatusForbidden, map[string]any{
			"status":  false,
			"message": "you cannot delete this comment",
		})
		return
	}

	if err != nil {
		log.Println(err)

		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "failed to delete comment",
		})
		return
	}

	for _, imagePath := range imagePaths {
		helpers.RemoveCommentMedia(imagePath)
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status":  true,
		"message": "comment deleted!",
	})
}

func (app *App) VoteGroupComment(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)

	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	commentID, err := strconv.Atoi(
		r.URL.Query().Get("commentId"),
	)

	if err != nil || commentID <= 0 {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid comment id",
		})
		return
	}

	vote, err := strconv.Atoi(
		r.URL.Query().Get("vote"),
	)

	if err != nil || (vote != 1 && vote != -1) {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid vote",
		})
		return
	}

	postID, err := groups.GroupCommentPostID(app.DB, commentID)

	if err == sql.ErrNoRows {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid comment id",
		})
		return
	}

	if err != nil {
		log.Println(err)

		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "failed to vote on comment",
		})
		return
	}

	allowed, err := groups.UserInPostGroup(app.DB, postID, userID)

	if err != nil {
		log.Println(err)

		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "failed to vote on comment",
		})
		return
	}

	if !allowed {
		helpers.WriteJson(w, http.StatusForbidden, map[string]any{
			"status":  false,
			"message": "you are not a member of this group",
		})
		return
	}

	err = groups.VoteGroupComment(
		app.DB,
		commentID,
		userID,
		vote,
	)

	if err != nil {
		log.Println(err)

		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "failed to vote on comment",
		})
		return
	}

	if vote == 1 {
		app.notifyGroupCommentLike(userID, commentID)
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status":  true,
		"message": "comment vote updated!",
	})
}
