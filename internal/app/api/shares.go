package api

import (
	"encoding/json"
	"net/http"
	"social/database/chats"
	"social/internal/helpers"
	"social/internal/models"
)

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

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		if !ok {
			helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
				"status":  false,
				"message": "invalid data",
			})
			return
		}
	}

	msgData := map[string]any{
		"type":   "message",
		"postID": req.PostID,
	}

	response, err := json.Marshal(msgData)
	if err != nil {
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not send message",
		})
		return
	}

	for _, id := range req.UserIds {
		msg := models.Message{
			Content: string(response),
		}

		msg.GroupID, err = chats.HasPrivateChat(app.DB, userID, id)
		if err != nil {
			helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
				"status":  false,
				"message": "could not get user Data",
			})
			return
		}

		canMessage, err := chats.CanSendMessage(app.DB, userID, id)
		if err != nil {
			helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
				"status":  false,
				"message": "could not get user data",
			})
			return
		}

		if !canMessage {
			helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
				"status":  false,
				"message": "could not make chat",
			})
			return
		}

		if msg.GroupID != -1 {
			err = chats.AddMessages(app.DB, msg.Content, userID, msg.GroupID)
			if err != nil {
				helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
					"status":  false,
					"message": "could not send chat",
				})
				return
			}
			app.sendToUsers(msg, msg.GroupID, userID)
		} else {
			groupID, err := chats.MakePrivateChat(app.DB, userID, id)
			if err != nil {
				helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
					"status":  false,
					"message": "could not make chat",
				})
				return
			}

			err = chats.AddMessages(app.DB, msg.Content, userID, msg.GroupID)
			if err != nil {
				helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
					"status":  false,
					"message": "could not send chat",
				})
				return
			}
			app.sendToUsers(msg, groupID, userID)
		}
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status":  true,
		"message": "post shared",
	})
}
