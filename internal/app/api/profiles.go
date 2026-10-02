package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"social/database/chats"
	"social/database/notifications"
	"social/database/profiles"
	"social/internal/helpers"
	"social/internal/models"
	"strconv"
)

func (app *App) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	requestedID := r.URL.Query().Get("id")
	profileID, err := strconv.Atoi(requestedID)
	if err != nil || profileID <= 0 {
		log.Println(err)
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid user id",
		})
		return
	}

	if profileID == userID {
		log.Println("same id")
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "cannot view your own profile this way",
		})
		return
	}

	isPrivate, err := profiles.IsPrivate(app.DB, profileID)
	if err != nil {
		if err == sql.ErrNoRows {
			helpers.WriteJson(w, http.StatusNotFound, map[string]any{
				"status":  false,
				"message": "no user found",
			})
			return
		}

		log.Println("here5", err)
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not check profile privacy",
		})
		return
	}

	isFollowing, err := profiles.CheckFollower(app.DB, userID, profileID)
	if err != nil {
		if err == sql.ErrNoRows {
			isFollowing = -1
		} else {
			log.Println("here3", err)
			helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
				"status":  false,
				"message": "could not check follow status",
			})
			return
		}
	}

	// Public profile OR accepted follower
	if !isPrivate || isFollowing == 1 {
		userData, err := profiles.GetUserData(app.DB, profileID)
		if err != nil {
			if err == sql.ErrNoRows {
				helpers.WriteJson(w, http.StatusNotFound, map[string]any{
					"status":       false,
					"showProfile":  false,
					"followStatus": -1,
					"message":      "no user found",
				})
				return
			}
			log.Println("here1", err)

			helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
				"status":       false,
				"showProfile":  false,
				"followStatus": -1,
				"message":      "could not get profile data",
			})
			return
		}

		canMessage, err := chats.CanSendMessage(app.DB, userID, profileID)
		if err != nil {
			helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
				"status":       false,
				"showProfile":  false,
				"followStatus": -1,
				"message":      "could not get profile data",
			})
			return
		}
		helpers.WriteJson(w, http.StatusOK, map[string]any{
			"status":       true,
			"showProfile":  true,
			"followStatus": isFollowing,
			"data":         userData,
			"canMessage": canMessage,
		})
		return
	}

	// Private profile and user is not following
	userData, err := profiles.GetPrivateProfileData(app.DB, profileID)
	if err != nil {
		if err == sql.ErrNoRows {
			helpers.WriteJson(w, http.StatusNotFound, map[string]any{
				"status":       false,
				"showProfile":  false,
				"followStatus": -1,
				"message":      "no user found",
			})
			return
		}

		log.Println("here0", err)
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":       false,
			"showProfile":  false,
			"followStatus": -1,
			"message":      "could not get profile data",
		})
		return
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status":       true,
		"showProfile":  false,
		"followStatus": isFollowing,
		"data":         userData,
	})
}

func (app *App) RequestFollow(w http.ResponseWriter, r *http.Request) {
	followerID, ok := r.Context().Value("userID").(int)

	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	queryID := r.URL.Query().Get("targetid")

	targetID, err := strconv.Atoi(queryID)

	if err != nil || targetID == followerID {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid target ID",
		})
		return
	}

	isPrivate, err := profiles.IsPrivate(app.DB, targetID)

	if err != nil {
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "failed to get user data",
		})
		return
	}

	var requestCode int

	if isPrivate {
		requestCode = 0
	} else {
		requestCode = 1
	}

	if err := profiles.SendFollowRequest(app.DB, targetID, followerID, requestCode); err != nil {
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "request to follow failed",
		})
		return
	}

	var notification models.NewNotification

	if requestCode == 1 {
		notification = models.NewNotification{
			UserID:       targetID,
			Message:      "You have a new follower",
			FollowUserID: &followerID,
		}
	} else {
		notification = models.NewNotification{
			UserID:              targetID,
			Message:             "You have a new follow request",
			FollowRequestUserID: &followerID,
		}
	}

	if err := notifications.InsertNotification(app.DB, notification); err != nil {
		log.Println("failed to create notification:", err)
	}

	notificationData, err := json.Marshal(notification)

	if err == nil {
		wsMessage, err := json.Marshal(models.WSPayload{
			Type: "notification",
			Data: notificationData,
		})

		if err == nil {
			println(wsMessage)
			// app.SendToUser(targetID, wsMessage)
		}
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status":       true,
		"followStatus": requestCode,
		"message":      "request sent",
	})
}

func (app *App) CancelRequest(w http.ResponseWriter, r *http.Request) {
	followerID, ok := r.Context().Value("userID").(int)
	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	queryID := r.URL.Query().Get("targetid")
	targetID, err := strconv.Atoi(queryID)
	if err != nil || targetID == followerID {
		log.Println(err, "here")
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid target ID",
		})
		return
	}

	if err := profiles.SendFollowRequest(app.DB, targetID, followerID, -1); err != nil {
		log.Println(err)
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid target ID",
		})
		return
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status":       true,
		"followStatus": -1,
		"message":      "request removed",
	})
	return
}

func (app *App) GetFollowers(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	queryID := r.URL.Query().Get("targetid")
	queryOffset := r.URL.Query().Get("offset")

	offset, err := strconv.Atoi(queryOffset)
	if err != nil || offset < 0 {
		log.Println(err)
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid offset",
		})
		return
	}

	targetID := userID

	if queryID != "" && queryID != "null" && queryID != "undefined" {
		targetID, err = strconv.Atoi(queryID)
		if err != nil || targetID <= 0 {
			log.Println(err)
			helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
				"status":  false,
				"message": "invalid target id",
			})
			return
		}
	}

	followers, err := profiles.GetFollowers(app.DB, targetID, 20, offset)
	if err != nil {
		if err == sql.ErrNoRows {
			helpers.WriteJson(w, http.StatusOK, map[string]any{
				"status": false,
				"data":   nil,
			})
			return
		}

		log.Println(err)
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not get followers",
		})
		return
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status": true,
		"data":   followers,
	})
}

func (app *App) GetFollowing(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	queryID := r.URL.Query().Get("targetid")
	queryOffset := r.URL.Query().Get("offset")

	offset, err := strconv.Atoi(queryOffset)
	if err != nil || offset < 0 {
		log.Println(err)
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid offset",
		})
		return
	}

	targetID := userID
	log.Println(queryID)
	if queryID != "" && queryID != "null" && queryID != "undefined" {
		targetID, err = strconv.Atoi(queryID)
		if err != nil || targetID <= 0 {
			log.Println(err)
			helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
				"status":  false,
				"message": "invalid target id",
			})
			return
		}
	}

	following, err := profiles.GetFollowing(app.DB, targetID, 20, offset)
	if err != nil {
		if err == sql.ErrNoRows {
			helpers.WriteJson(w, http.StatusOK, map[string]any{
				"status": false,
				"data":   nil,
			})
			return
		}

		log.Println(err)
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not get following",
		})
		return
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status": true,
		"data":   following,
	})
}

func (app *App) RemoveFollower(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	queryID := r.URL.Query().Get("targetid")
	followerID, err := strconv.Atoi(queryID)
	if err != nil || followerID == userID {
		log.Println(err)
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid target ID",
		})
		return
	}

	// The requesting user is the target (owner of the followers list);
	// the follower being removed is the one following them.
	if err := profiles.SendFollowRequest(app.DB, userID, followerID, -1); err != nil {
		log.Println(err)
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not remove follower",
		})
		return
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status":  true,
		"message": "follower removed",
	})
}

func (app *App) AcceptFollowRequest(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)

	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	queryID := r.URL.Query().Get("targetid")

	requesterID, err := strconv.Atoi(queryID)

	if err != nil || requesterID == userID {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid target ID",
		})
		return
	}

	err = profiles.AcceptFollowRequest(app.DB, requesterID, userID)

	if err != nil {
		log.Println(err)

		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not accept follow request",
		})
		return
	}

	notification := models.NewNotification{
		UserID:                    requesterID,
		Message:                   "Your follow request was accepted",
		FollowRequestAcceptUserID: &userID,
	}

	if err := notifications.InsertNotification(app.DB, notification); err != nil {
		log.Println("failed to create notification:", err)
	}

	notificationData, err := json.Marshal(notification)

	if err == nil {
		wsMessage, err := json.Marshal(models.WSPayload{
			Type: "notification",
			Data: notificationData,
		})

		if err == nil {
			println(wsMessage)
			// app.SendToUser(requesterID, wsMessage)
		}
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status":  true,
		"message": "follow request accepted",
	})
}
