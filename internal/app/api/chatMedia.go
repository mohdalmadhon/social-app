package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"social/database/chats"
	"social/database/users"
	"social/internal/helpers"
	"social/internal/models"
	"strconv"
)

func (app *App) SendChatMedia(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, helpers.MaxChatMediaSize+(1<<20))

	if err := r.ParseMultipartForm(helpers.MaxChatMediaSize); err != nil {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid form data or file is too large",
		})
		return
	}

	groupID, err := strconv.Atoi(r.FormValue("groupID"))
	if err != nil {
		groupID = -1
	}

	targetID, err := strconv.Atoi(r.FormValue("userID"))
	if err != nil {
		targetID = -1
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "image is required",
		})
		return
	}
	defer file.Close()

	if groupID <= 0 {
		if targetID <= 0 || targetID == userID {
			helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
				"status":  false,
				"message": "invalid user",
			})
			return
		}
	} else {
		inGroup, err := chats.UserInGroup(app.DB, userID, groupID)
		if err != nil {
			log.Println(err)
			helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
				"status":  false,
				"message": "could not verify chat",
			})
			return
		}

		if !inGroup {
			helpers.WriteJson(w, http.StatusForbidden, map[string]any{
				"status":  false,
				"message": "could not send message",
			})
			return
		}

		isPrivate, otherID, err := chats.IsPrivateChat(app.DB, groupID, userID)
		if err != nil && err != sql.ErrNoRows {
			log.Println(err)
			helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
				"status":  false,
				"message": "could not verify chat",
			})
			return
		}

		if isPrivate {
			targetID = otherID
		} else {
			targetID = -1
		}
	}

	if targetID > 0 {
		canMessage, err := chats.CanSendMessage(app.DB, userID, targetID)
		if err != nil {
			log.Println(err)
			helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
				"status":  false,
				"message": "could not verify message permission",
			})
			return
		}

		if !canMessage {
			helpers.WriteJson(w, http.StatusForbidden, map[string]any{
				"status":  false,
				"message": "could not send message because of user preference",
			})
			return
		}
	}

	path, kind, err := helpers.SaveChatMedia(file, header)
	if err != nil {
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
		return
	}

	if groupID <= 0 {
		existingGroupID, err := chats.HasPrivateChat(app.DB, userID, targetID)
		if err != nil {
			log.Println(err)
			os.Remove(filepath.Join("uploads", path))
			helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
				"status":  false,
				"message": "could not verify chat",
			})
			return
		}

		if existingGroupID != -1 {
			groupID = existingGroupID
		} else {
			groupID, err = chats.MakePrivateChat(app.DB, userID, targetID)
			if err != nil {
				log.Println(err)
				os.Remove(filepath.Join("uploads", path))
				helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
					"status":  false,
					"message": "could not make chat",
				})
				return
			}
		}
	}

	content, err := json.Marshal(map[string]string{
		"type": kind,
		"path": path,
	})
	if err != nil {
		log.Println(err)
		os.Remove(filepath.Join("uploads", path))
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not send image",
		})
		return
	}

	if err := chats.AddMessages(app.DB, string(content), userID, groupID); err != nil {
		log.Println(err)
		os.Remove(filepath.Join("uploads", path))
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not send image",
		})
		return
	}

	sender, err := users.GetUserSimpleData(app.DB, userID)
	if err == nil {
		app.sendToUsers(models.Message{
			Content: string(content),
			Sender: models.UserRegistration{
				ID:        userID,
				FirstName: sender.FirstName,
				LastName:  sender.LastName,
				Avatar:    sender.Avatar,
			},
			GroupID: groupID,
		}, groupID, userID)
	} else {
		log.Println("get sender error:", err)
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status":  true,
		"groupID": groupID,
		"content": string(content),
	})
}
