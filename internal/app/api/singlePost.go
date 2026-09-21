package api

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"social/database/posts"
	"social/internal/helpers"
)

func (app *App) GetSinglePost(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)

	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	postID, err := strconv.Atoi(r.URL.Query().Get("postID"))

	if err != nil || postID <= 0 {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid post id",
		})
		return
	}

	post, err := posts.GetSinglePost(app.DB, postID, userID)

	if err == sql.ErrNoRows {
		helpers.WriteJson(w, http.StatusNotFound, map[string]any{
			"status":  false,
			"message": "post not found",
		})
		return
	}

	if err != nil {
		log.Println(err)

		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not get post",
		})
		return
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status": true,
		"userId": userID,
		"data":   post,
	})
}
