package api

import (
	"log"
	"net/http"
	"strconv"

	"social/database/groups"
	"social/internal/helpers"
)

func (app *App) GetGroupPosts(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)

	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	groupID, err := strconv.Atoi(
		r.URL.Query().Get("groupID"),
	)

	if err != nil || groupID <= 0 {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid group",
		})
		return
	}

	offset := 0

	offsetValue := r.URL.Query().Get("offset")

	if offsetValue != "" {
		offset, err = strconv.Atoi(offsetValue)

		if err != nil || offset < 0 {
			helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
				"status":  false,
				"message": "invalid offset",
			})
			return
		}
	}

	userIN, err := groups.UserIN(app.DB, groupID, userID)

	if err != nil {
		log.Println(err)

		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not get group posts",
		})
		return
	}

	if !userIN {
		helpers.WriteJson(w, http.StatusForbidden, map[string]any{
			"status":  false,
			"message": "you are not a member of this group",
		})
		return
	}

	posts, err := groups.GetGroupFeedPosts(app.DB, groupID, userID, 10, offset)

	if err != nil {
		log.Println(err)

		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not get group posts",
		})
		return
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status": true,
		"userId": userID,
		"data":   posts,
	})
}
