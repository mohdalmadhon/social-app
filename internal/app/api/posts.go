package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"social/database/posts"
	"social/internal/helpers"
	"social/internal/models"
	"social/internal/validation"
)

func (app *App) AddPost(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "could not read form",
		})
		return
	}

	allowComments, err := strconv.Atoi(r.FormValue("allowComments"))
	if err != nil {
		allowComments = 1
	}

	groupID, err := strconv.Atoi(r.FormValue("groupID"))
	if err != nil {
		groupID = 0
	}

	var taggedPeople []int

	taggedPeopleData := r.FormValue("taggedPeople")

	if taggedPeopleData != "" {
		if err := json.Unmarshal([]byte(taggedPeopleData), &taggedPeople); err != nil {
			helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
				"status":  false,
				"message": "invalid tagged people",
			})
			return
		}
	}

	post := models.RegsiterPost{
		UserID:        userID,
		Content:       r.FormValue("content"),
		AllowComments: allowComments,
		GroupID:       groupID,
		Location:      r.FormValue("location"),
		PeopleTagged:  taggedPeople,
	}

	file, header, err := r.FormFile("image")
	
	res := validation.ValidatePost(&post, header)
	if res.Field != "" {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "could not upload post",
		})
		return
	}

	if err == nil {
		defer file.Close()

		imagePath, err := helpers.SaveUploads(file, header, "post")
		if err != nil {
			helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
				"status":  false,
				"message": "could not save image",
			})
			return
		}

		post.Image_path = imagePath
	} else if err != http.ErrMissingFile {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid image",
		})
		return
	}

	if err := posts.GroupExists(app.DB, post.GroupID, userID); err != nil {
		if err == sql.ErrNoRows {
			helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
				"status":  false,
				"message": "group does not exists",
			})
			return
		}
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not get group post",
		})
		return
	}

	postID, err := posts.AddPost(app.DB, post)

	if err != nil {
		log.Println("here2", err)
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not upload post",
		})
		return
	}

	actorName := app.actorName(userID)
	notified := map[int]bool{userID: true}

	for _, taggedUserID := range taggedPeople {
		if notified[taggedUserID] {
			continue
		}

		notified[taggedUserID] = true

		app.notify(userID, models.NewNotification{
			UserID:            taggedUserID,
			Message:           fmt.Sprintf("%s tagged you in a post", actorName),
			PostIDTag:         &postID,
			PostMentionUserID: &userID,
		})
	}

	for _, mentionedID := range app.mentionedUserIDs(post.Content, notified) {
		notified[mentionedID] = true

		app.notify(userID, models.NewNotification{
			UserID:            mentionedID,
			Message:           fmt.Sprintf("%s mentioned you in a post", actorName),
			PostIDTag:         &postID,
			PostMentionUserID: &userID,
		})
	}

	helpers.WriteJson(w, http.StatusCreated, map[string]any{
		"status":  true,
		"message": "post uploaded successfully",
	})
}

func (app *App) GetHomePosts(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)

	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	offset := 0

	if value := r.URL.Query().Get("offset"); value != "" {
		var err error

		offset, err = strconv.Atoi(value)

		if err != nil || offset < 0 {
			helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
				"status":  false,
				"message": "invalid offset",
			})
			return
		}
	}

	posts, err := posts.GetHomePosts(app.DB, userID, offset)

	if err != nil {
		log.Println(err)
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not get home posts",
		})
		return
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status": true,
		"posts":  posts,
	})
}

func (app *App) PostReaction(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	var rect models.Reaction
	if err := json.NewDecoder(r.Body).Decode(&rect); err != nil {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid reaction ",
		})
		return
	}

	if rect.Value != 1 && rect.Value != -1 {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid reaction",
		})
		return
	}

	rect.UserID = userID
	err, deletion := posts.InsertReaction(app.DB, rect)
	if err != nil {

		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not insert reaction",
		})
		return
	}

	if deletion {
		helpers.WriteJson(w, http.StatusOK, map[string]any{
			"status":  true,
			"message": "reaction inserted!",
		})
		return
	}
	// send notification (best effort: the reaction itself is already saved)

	targetID, err := posts.GetPostOwnerID(app.DB, rect.PostID)
	if err != nil {
		log.Println("could not find post owner:", err)
	} else {
		notification := models.NewNotification{
			UserID:    targetID,
			PostIDTag: &rect.PostID,
		}

		name := app.actorName(userID)

		if rect.Value == 1 {
			notification.PostLikeUserID = &userID
			notification.Message = fmt.Sprintf("%s liked your post", name)
		} else {
			notification.PostDislikeUserID = &userID
			notification.Message = fmt.Sprintf("%s disliked your post", name)
		}

		app.notify(userID, notification)
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status":  true,
		"message": "reaction inserted!",
	})
}

func (app *App) GetUserPosts(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)

	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	targetIDStr := r.URL.Query().Get("targetID")
	targetID := 0
	log.Println(targetIDStr, "str")

	if targetIDStr != "" {
		var err error

		targetID, err = strconv.Atoi(targetIDStr)
		if err != nil {
			helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
				"status":  false,
				"message": "invalid targetID",
			})
			return
		}
	} else {
		log.Println("dont need filtering")
		targetID = userID
	}

	offsetStr := r.URL.Query().Get("offset")

	offset := 0

	if offsetStr != "" {
		var err error

		offset, err = strconv.Atoi(offsetStr)

		if err != nil {
			helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
				"status":  false,
				"message": "invalid offset",
			})
			return
		}
	}

	userPosts, err := posts.GetUserPosts(app.DB, targetID, offset)
	if err != nil {
		if err == sql.ErrNoRows {
			helpers.WriteJson(w, http.StatusOK, map[string]any{
				"status":  true,
				"message": "no posts",
			})
			return
		}

		log.Println(err)
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not get posts",
		})
		return
	}

	if targetID != userID {
		log.Println("need filtering")
		userPosts, err = posts.FilterPosts(app.DB, &userPosts, userID, targetID)

		if err != nil {
			log.Println(err, "filtering")
			helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
				"status":  false,
				"message": "could not filter posts",
			})
			return
		}
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status": true,
		"data":   userPosts,
	})
}

func (app *App) DeletePost(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	postID, err := strconv.Atoi(r.URL.Query().Get("postId"))
	if err != nil {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid post id",
		})
		return
	}

	err = posts.DeletePost(app.DB, postID, userID)

	if err == sql.ErrNoRows {
		helpers.WriteJson(w, http.StatusForbidden, map[string]any{
			"status":  false,
			"message": "you cannot delete this post",
		})
		return
	}

	if err != nil {
		log.Println(err)
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not delete post",
		})
		return
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status":  true,
		"message": "post deleted",
	})
}

func (app *App) ViewPost(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	var postID int
	if err := json.NewDecoder(r.Body).Decode(&postID); err != nil {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid post id",
		})
		return
	}

	if err := posts.ViewPost(app.DB, postID, userID); err != nil && err.Error() != "UNIQUE constraint failed: post_views.user_id, post_views.post_id" {
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not add post",
		})
		return
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status":  true,
		"message": "all good",
	})
}
