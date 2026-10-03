package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"social/database/posts"
	"social/internal/helpers"
	"social/internal/models"
	"social/internal/validation"
)

func (app *App) AddComment(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)

	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	var comment models.Comment

	if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid comment",
		})
		return
	}

	comment.User.ID = userID

	if err := validation.ValidateComment(comment); err != nil {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": err.Error(),
		})
		return
	}

	comment, err := posts.InsertComment(app.DB, comment)

	if err != nil {
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "failed to insert comment",
		})
		return
	}

	app.notifyComment(userID, comment)

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status":  true,
		"message": "comment inserted!",
		"comment": comment,
	})
}

func (app *App) DeleteComment(w http.ResponseWriter, r *http.Request) {
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

	if err != nil {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid comment id",
		})
		return
	}

	err = posts.DeleteComment(
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
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "failed to delete comment",
		})
		return
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status":  true,
		"message": "comment deleted!",
	})
}

func (app *App) GetComments(w http.ResponseWriter, r *http.Request) {
	_, ok := r.Context().Value("userID").(int)

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

	if err != nil {
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

	comments, err := posts.GetComments(
		app.DB,
		postID,
		replyTo,
		50,
		0,
		"latest",
	)

	if err != nil {
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "failed to get comments",
		})
		return
	}

	if comments == nil {
		comments = []models.Comment{}
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status":   true,
		"comments": comments,
	})
}

func (app *App) VoteComment(w http.ResponseWriter, r *http.Request) {
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

	if err != nil {
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

	err = posts.VoteComment(
		app.DB,
		commentID,
		userID,
		vote,
	)

	if err != nil {
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "failed to vote on comment",
		})
		return
	}

	if vote == 1 {
		app.notifyCommentLike(userID, commentID)
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status":  true,
		"message": "comment vote updated!",
	})
}