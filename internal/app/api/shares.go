package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"social/database/chats"
	"social/database/profiles"
	"social/database/users"
	"social/internal/helpers"
	"social/internal/models"
	"social/internal/validation"
)

const maxShareRecipients = 30

func (app *App) shareToUsers(userID int, targetIDs []int, content string) (int, error) {
	sender, err := users.GetUserSimpleData(app.DB, userID)
	if err != nil {
		return 0, err
	}

	if len(targetIDs) > maxShareRecipients {
		targetIDs = targetIDs[:maxShareRecipients]
	}

	seen := make(map[int]bool, len(targetIDs))
	sent := 0

	for _, id := range targetIDs {
		if id <= 0 || id == userID || seen[id] {
			continue
		}

		seen[id] = true

		canMessage, err := chats.CanSendMessage(app.DB, userID, id)
		if err != nil {
			return sent, err
		}

		if !canMessage {
			continue
		}

		groupID, err := chats.HasPrivateChat(app.DB, userID, id)
		if err != nil {
			return sent, err
		}

		if groupID == -1 {
			groupID, err = chats.MakePrivateChat(app.DB, userID, id)
			if err != nil {
				return sent, err
			}
		}

		if err := chats.AddMessages(app.DB, content, userID, groupID); err != nil {
			return sent, err
		}

		app.sendToUsers(models.Message{
			Content: content,
			Sender: models.UserRegistration{
				ID:        userID,
				FirstName: sender.FirstName,
				LastName:  sender.LastName,
				Avatar:    sender.Avatar,
			},
			GroupID: groupID,
		}, groupID, userID)

		sent++
	}

	return sent, nil
}

func (app *App) SharePost(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	var req struct {
		UserIds []int `json:"userIds"`
		PostID  int   `json:"postID"`
	}

	r.Body = http.MaxBytesReader(w, r.Body, validation.MaxJSONBody)

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.PostID <= 0 || len(req.UserIds) == 0 {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid data",
		})
		return
	}

	if !readIDList(w, "share list", req.UserIds) {
		return
	}

	content, err := json.Marshal(map[string]any{
		"type":   "message",
		"postID": req.PostID,
	})
	if err != nil {
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not send message",
		})
		return
	}

	sent, err := app.shareToUsers(userID, req.UserIds, string(content))
	if err != nil {
		log.Println("share post error:", err)
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not share post",
		})
		return
	}

	if sent == 0 {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "could not message any of the selected users",
		})
		return
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status":  true,
		"message": "post shared",
		"sent":    sent,
	})
}

func (app *App) ShareProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	var req struct {
		UserIds   []int `json:"userIds"`
		ProfileID int   `json:"profileID"`
	}

	r.Body = http.MaxBytesReader(w, r.Body, validation.MaxJSONBody)

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid data",
		})
		return
	}

	if len(req.UserIds) == 0 {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid data",
		})
		return
	}

	if !readIDList(w, "share list", req.UserIds) {
		return
	}

	if req.ProfileID == -99 {
		req.ProfileID = userID
	} else if req.ProfileID <= 0 {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid profile ID",
		})
		return
	}

	card, err := profiles.GetShareProfileCard(app.DB, req.ProfileID)
	if err != nil {
		if err == sql.ErrNoRows {
			helpers.WriteJson(w, http.StatusNotFound, map[string]any{
				"status":  false,
				"message": "no user found",
			})
			return
		}

		log.Println("share profile lookup error:", err)

		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not get profile data",
		})
		return
	}

	content, err := json.Marshal(map[string]any{
		"type":      "profile",
		"profileID": card.ID,
		"firstName": card.FirstName,
		"lastName":  card.LastName,
		"username":  card.UserName,
		"avatar":    card.Avatar,
	})

	if err != nil {
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not send message",
		})
		return
	}

	sent, err := app.shareToUsers(userID, req.UserIds, string(content))
	if err != nil {
		log.Println("share profile error:", err)

		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not share profile",
		})
		return
	}

	if sent == 0 {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "could not message any of the selected users",
		})
		return
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status":  true,
		"message": "profile shared",
		"sent":    sent,
	})
}
