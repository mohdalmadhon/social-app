package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"social/database/posts"
	"social/database/profiles"
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

	r.Body = http.MaxBytesReader(w, r.Body, validation.MaxPostMediaSize+(1<<20))

	if err := r.ParseMultipartForm(validation.MaxFormMemory); err != nil {
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
			"field":   res.Field,
			"message": res.Message,
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

	if r.URL.Query().Get("type") == "videos" {
		limit := 5

		if value := r.URL.Query().Get("limit"); value != "" {
			var err error

			limit, err = strconv.Atoi(value)

			if err != nil || limit < 1 {
				helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
					"status":  false,
					"message": "invalid limit",
				})
				return
			}

			if limit > 50 {
				limit = 50
			}
		}

		videos, err := posts.GetHomeVideos(app.DB, userID, offset, limit)

		if err != nil {
			log.Println(err)
			helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
				"status":  false,
				"message": "could not get home videos",
			})
			return
		}

		helpers.WriteJson(w, http.StatusOK, map[string]any{
			"status":  true,
			"posts":   videos,
			"hasMore": len(videos) == limit,
		})
		return
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
	r.Body = http.MaxBytesReader(w, r.Body, validation.MaxJSONBody)

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

		if err != nil || offset < 0 {
			helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
				"status":  false,
				"message": "invalid offset",
			})
			return
		}
	}

	limit := 9
	limitStr := r.URL.Query().Get("limit")

	if limitStr != "" {
		var err error

		limit, err = strconv.Atoi(limitStr)

		if err != nil || limit < 1 {
			helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
				"status":  false,
				"message": "invalid limit",
			})
			return
		}

		if limit > 50 {
			limit = 50
		}
	}

	postType := r.URL.Query().Get("type")
	videosOnly := postType == "videos"

	if postType == "tagged" {
		app.writeTaggedPosts(w, userID, targetID, offset, limit)
		return
	}

	userPosts, err := posts.GetUserPosts(app.DB, targetID, offset, limit, videosOnly)
	if err != nil {
		if err == sql.ErrNoRows {
			helpers.WriteJson(w, http.StatusOK, map[string]any{
				"status":  true,
				"message": "no posts",
				"hasMore": false,
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

	hasMore := len(userPosts) == limit

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
		"status":  true,
		"data":    userPosts,
		"hasMore": hasMore,
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
	r.Body = http.MaxBytesReader(w, r.Body, validation.MaxJSONBody)

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

func (app *App) writeTaggedPosts(w http.ResponseWriter, userID, targetID, offset, limit int) {
	if targetID != userID {
		isPrivate, err := posts.IsPrivateProfile(app.DB, targetID)
		if err != nil {
			log.Println(err)
			helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
				"status":  false,
				"message": "could not get posts",
			})
			return
		}

		if isPrivate {
			status, err := profiles.CheckFollower(app.DB, userID, targetID)
			if err != nil && err != sql.ErrNoRows {
				log.Println(err)
				helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
					"status":  false,
					"message": "could not get posts",
				})
				return
			}

			if err == sql.ErrNoRows || status != 1 {
				helpers.WriteJson(w, http.StatusOK, map[string]any{
					"status":  true,
					"data":    []models.Post{},
					"hasMore": false,
				})
				return
			}
		}
	}

	taggedPosts, err := posts.GetTaggedPosts(app.DB, targetID, offset, limit)
	if err != nil {
		log.Println(err)
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not get posts",
		})
		return
	}

	hasMore := len(taggedPosts) == limit

	visiblePosts, err := posts.FilterTaggedPosts(app.DB, taggedPosts, userID, targetID)
	if err != nil {
		log.Println(err)
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not filter posts",
		})
		return
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status":  true,
		"data":    visiblePosts,
		"hasMore": hasMore,
	})
}
