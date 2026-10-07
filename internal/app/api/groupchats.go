package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"social/database/chats"
	"social/database/groups"
	"social/database/notifications"
	"social/database/users"
	"social/internal/helpers"
	"social/internal/models"
	"social/internal/validation"
	"strconv"
)

func (app *App) GetGroups(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	privatChatQuery := r.URL.Query().Get("private")
	isPrivate, err := strconv.Atoi(privatChatQuery)
	if err != nil || (isPrivate != 0 && isPrivate != 1) {
		helpers.WriteJson(w, http.StatusBadGateway, map[string]any{
			"status":  false,
			"message": "what you want to get excatly? specify please.........  bro....",
		})
		return
	}

	offset := 0
	offsetQuery := r.URL.Query().Get("offset")
	if offsetQuery != "" && offsetQuery != "null" && offsetQuery != "undifiend" {
		var err error
		offset, err = strconv.Atoi(offsetQuery)
		if err != nil {
			helpers.WriteJson(w, http.StatusBadGateway, map[string]any{
				"status":  false,
				"message": "invalid offset",
			})
			return
		}
	}

	if isPrivate == 1 {
		chats, err := chats.GetPrivateChatsList(app.DB, userID, offset)
		if err != nil && err != sql.ErrNoRows {
			helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
				"status":  false,
				"message": "failed to get users list",
			})
			return
		}

		helpers.WriteJson(w, http.StatusOK, map[string]any{
			"status": true,
			"data":   chats,
		})
	}
	if isPrivate == 0 {
		groups, err := groups.GetGroupChats(app.DB, userID, offset)
		if err != nil {
			helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
				"status":  true,
				"message": err.Error(),
			})
			return
		}

		helpers.WriteJson(w, http.StatusOK, map[string]any{
			"status": true,
			"data":   groups,
		})
		return
	}

}

func (app *App) SearchPrivateChats(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}
	privatChatQuery := r.URL.Query().Get("private")
	isPrivate, err := strconv.Atoi(privatChatQuery)
	if err != nil || (isPrivate != 0 && isPrivate != 1) {
		helpers.WriteJson(w, http.StatusBadGateway, map[string]any{
			"status":  false,
			"message": "what you want to get excatly? specify please.........  bro....",
		})
		return
	}

	offset := 0
	offsetQuery := r.URL.Query().Get("offset")
	if offsetQuery != "" && offsetQuery != "null" && offsetQuery != "undifiend" {
		var err error
		offset, err = strconv.Atoi(offsetQuery)
		if err != nil {
			helpers.WriteJson(w, http.StatusBadGateway, map[string]any{
				"status":  false,
				"message": "invalid offset",
			})
			return
		}
	}

	search := r.URL.Query().Get("search")
	if search == "" {
		helpers.WriteJson(w, http.StatusOK, map[string]any{
			"status": true,
			"data":   nil,
		})
		return
	}

	if isPrivate == 1 {
		chats, err := chats.SearchChatUsers(app.DB, userID, offset, search)
		if err != nil && err != sql.ErrNoRows {
			log.Println(err)
			helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
				"status":  false,
				"message": "failed to get chats",
			})
			return
		}

		helpers.WriteJson(w, http.StatusOK, map[string]any{
			"status": true,
			"data":   chats,
		})
	}
}

func (app *App) AddMessages(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)

	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	type request struct {
		Offset  int    `json:"offset"`
		GroupID int    `json:"groupID"`
		UserID  int    `json:"userID"`
		Content string `json:"content"`
	}

	var req request

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println(err)
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid data",
		})
		return
	}

	if req.Content == "" {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "message cannot be empty",
		})
		return
	}

	groupID := req.GroupID

	if groupID <= 0 {
		if req.UserID <= 0 || req.UserID == userID {
			helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
				"status":  false,
				"message": "invalid user",
			})
			return
		}

		canMessage, err := chats.CanSendMessage(app.DB, userID, req.UserID)
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

		existingGroupID, err := chats.HasPrivateChat(app.DB, userID, req.UserID)
		if err != nil {
			helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
				"status":  false,
				"message": "could not verify chat",
			})
			return
		}

		if existingGroupID != -1 {
			groupID = existingGroupID
		} else {

			groupID, err = chats.MakePrivateChat(app.DB, userID, req.UserID)
			if err != nil {
				helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
					"status":  false,
					"message": "could not make chat",
				})
				return
			}
		}
	} else {
		log.Println(req.GroupID)
		exists, err := groups.GroupExists(app.DB, req.GroupID)
		if err != nil {
			helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
				"status":  false,
				"message": "group does not exists",
			})
			return
		}

		userIN, err := groups.UserIN(app.DB, req.GroupID, userID)
		if err != nil {
			helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
				"status":  false,
				"message": "bad request",
			})
			return
		}

		if !exists || !userIN {
			helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
				"status":  false,
				"message": "could not send message",
			})
			return
		}

		isPrivate, targetID, err := chats.IsPrivateChat(app.DB, groupID, userID)
		if err != nil {
			log.Println(err)
			helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
				"status":  false,
				"message": "could not verify chat",
			})
			return
		}

		if isPrivate {
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
	}

	err := chats.AddMessages(app.DB, req.Content, userID, groupID)
	if err != nil {
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not send chat",
		})
		return
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status":  true,
		"message": "message sent",
		"groupID": groupID,
	})
}

func (app *App) GetMessages(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	groupID, err := strconv.Atoi(r.URL.Query().Get("groupID"))
	if err != nil {
		log.Println(err)
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid groupID",
		})
		return
	}

	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil {
		log.Println(err)
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid groupID",
		})
		return
	}

	exists, err := chats.ChatExists(app.DB, groupID)
	if err != nil {
		log.Println(err, "here1")
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not verify chat",
		})
		return
	}

	if exists {
		userIN, err := chats.UserInGroup(app.DB, userID, groupID)
		if err != nil {
			log.Println(err, "here2")
			helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
				"status":  false,
				"message": "could not verify group",
			})
			return
		}

		if !userIN {
			log.Println("user not in")
			helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
				"status":  false,
				"message": "invalid request",
			})
			return
		}
	}

	msgs, err := chats.GetMessages(app.DB, userID, groupID, offset)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err, "here3")
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not get messages",
		})
		return
	}

	isPrivate, targetID, err := chats.IsPrivateChat(app.DB, groupID, userID)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not verify chat",
		})
		return
	}

	if isPrivate && targetID > 0 || (groupID == 0) {
		targetID, err := strconv.Atoi(r.URL.Query().Get("userID"))
		if err != nil {
			helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
				"status":  false,
				"message": "could not get chat data",
			})
			return
		}
		canMessage, err := chats.CanSendMessage(app.DB, userID, targetID)
		if err != nil {
			helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
				"status":  false,
				"message": "could not get group data",
			})
			return
		}

		helpers.WriteJson(w, http.StatusOK, map[string]any{
			"status":     true,
			"data":       msgs,
			"canMessage": canMessage,
		})
		return
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status": true,
		"data":   msgs,
	})
}

func (app *App) MakeNewGroup(w http.ResponseWriter, r *http.Request) {
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
			"message": "invalid form data",
		})
		return
	}

	group := models.Group{
		Title:       r.FormValue("title"),
		Description: r.FormValue("description"),
		UserID:      userID,
	}

	var userIDs []int

	usersArray := r.FormValue("users")
	log.Println(usersArray)
	if usersArray != "" {
		if err := json.Unmarshal([]byte(usersArray), &userIDs); err != nil {
			helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
				"status":  false,
				"message": "invalid users array",
			})
			return
		}
	}

	for _, id := range userIDs {
		if id == userID {
			continue
		}

		_, allowed, err := app.groupRelation(userID, id)

		if err != nil {
			log.Println(err)
			helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
				"status":  false,
				"message": "failed to check data",
			})
			return
		}

		if !allowed {
			helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
				"status":  false,
				"message": "you can only add friends, followers or people you follow",
			})
			return
		}
	}

	if group.UserID != 0 {
		userIDs = append(userIDs, group.UserID)
	}

	avatar, header, err := r.FormFile("avatar")

	if err := validation.ValidateGroup(group, header); err != nil {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": err.Error(),
		})
		return
	}

	if err == nil {
		defer avatar.Close()

		path, err := helpers.SaveUploads(avatar, header, "group/avatar")
		if err != nil {
			log.Println(err)
			helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
				"status":  false,
				"message": "failed to save group avatar",
			})
			return
		}

		group.Avatar = path
	}

	g, ids, err := groups.MakeNewGroup(app.DB, group, userIDs)
	if err != nil {
		log.Println(err)

		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "failed to make new group",
		})
		return
	}

	requests := []int{}

	for _, id := range ids {
		if id == userID {
			continue
		}

		result, err := app.addOrInvite(userID, id, g.ID, g.Title)

		if err != nil {
			log.Println(err)
			continue
		}

		if result == memberRequested {
			requests = append(requests, id)
		}
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status":  true,
		"message": "group created",
		"request": map[string]any{
			"usersIds":  requests,
			"groupData": g,
		},
	})

}

func (app *App) SearchInvites(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	var groupID int
	search := r.URL.Query().Get("search")

	groupIDStr := r.URL.Query().Get("groupID")
	if groupIDStr != "" {
		var err error
		groupID, err = strconv.Atoi(groupIDStr)
		if err != nil {
			log.Println(err)
			helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
				"status":  false,
				"message": "invalid groupID",
			})
			return
		}
	} else {
		groupID = -1
	}

	users, err := groups.SearchInvites(app.DB, userID, groupID, search)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not get users",
		})
		return
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status": true,
		"data":   users,
	})
}

func (app *App) AcceptInvite(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	type Request struct {
		Status   int    `json:"status"`
		GroupID  int    `json:"groupID"`
		Content  string `json:"content"`
		SenderID int    `json:"senderID"`
	}

	var req Request

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "bad data",
		})
		return
	}

	fmt.Println(req.Content)
	if req.Status != 1 && req.Status != -1 {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid status",
		})
		return
	}

	hasInvite, err := notifications.HasGroupInvite(app.DB, userID, req.GroupID)

	if err != nil {
		log.Println(err)
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not check invite",
		})
		return
	}

	if !hasInvite {
		helpers.WriteJson(w, http.StatusNotFound, map[string]any{
			"status":  false,
			"message": "invite not found",
		})
		return
	}

	inviterID, err := groups.GetPendingInviter(app.DB, req.GroupID, userID)

	if err != nil {
		log.Println(err)
	}

	if err := groups.ChangeStatus(app.DB, userID, req.Status, req.GroupID); err != nil {
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not update group status",
		})
		return
	}

	if err := notifications.DeleteGroupInvites(app.DB, userID, req.GroupID); err != nil {
		log.Println(err)
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not delete invite",
		})
		return
	}

	if err := groups.DeleteInvite(app.DB, userID, req.SenderID, req.GroupID); err != nil {
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not delete invite",
		})
		return
	}

	if req.Status == 1 {
		app.notifyGroupInviteAccepted(userID, req.GroupID, inviterID)
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status":  true,
		"message": "invite removed",
	})

}

func (app *App) DiscoverGroups(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid offset",
		})
		return
	}

	search := r.URL.Query().Get("search")

	groups, err := groups.DiscoverGroups(app.DB, userID, offset, search)
	if err != nil && err != sql.ErrNoRows {
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not get groups",
		})
		return
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status": true,
		"data":   groups,
	})
	return
}

func (app *App) GetGroup(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	groupID, err := strconv.Atoi(r.URL.Query().Get("groupID"))
	if err != nil {
		log.Println(err)
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid groupID",
		})
		return
	}
	log.Println(groupID)
	group, err := groups.GetGroup(app.DB, userID, groupID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println(err)
			helpers.WriteJson(w, http.StatusOK, map[string]any{
				"status":  false,
				"message": "group does not exists",
			})
			return
		}

		log.Println(err)
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not get group data",
		})
		return
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status":  true,
		"data":    group,
		"isOwner": group.UserID == userID,
	})
}

func (app *App) SearchMembers(w http.ResponseWriter, r *http.Request) {
	_, ok := r.Context().Value("userID").(int)
	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	groupID, err := strconv.Atoi(r.URL.Query().Get("groupID"))
	if err != nil {
		log.Println(err)
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid groupID",
		})
		return
	}

	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil {
		log.Println(err)
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid offset",
		})
		return
	}

	search := r.URL.Query().Get("search")

	ids, err := groups.SearchGroupMembers(app.DB, groupID, search, offset, 20)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not get members",
		})
		return
	}

	var usersArray []models.UserRegistration
	for _, id := range ids {
		u, err := users.GetUserSimpleData(app.DB, id)
		if err != nil {
			helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
				"status":  false,
				"message": "could not get members",
			})
			return
		}

		usersArray = append(usersArray, u)
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status": true,
		"data":   usersArray,
	})
}

func (app *App) GroupRequest(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	type Request struct {
		GroupID int `json:"groupID"`
		Code    int `json:"code"`
	}

	var req Request

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid data",
		})
		return
	}

	if req.GroupID <= 0 {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid group",
		})
		return
	}

	if req.Code != 0 && req.Code != -1 {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid request code",
		})
		return
	}

	ownerID, err := groups.GetGroupOwner(app.DB, req.GroupID)
	if err != nil {
		if err == sql.ErrNoRows {
			helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
				"status":  false,
				"message": "could not find group",
			})
			return
		}

		log.Println(err)
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not get group data",
		})
		return
	}

	if req.Code == 0 {
		banned, err := groups.IsBanned(app.DB, req.GroupID, userID)
		if err != nil {
			log.Println(err)
			helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
				"status":  false,
				"message": "could not check group access",
			})
			return
		}

		if banned {
			helpers.WriteJson(w, http.StatusForbidden, map[string]any{
				"status":  false,
				"message": "you were removed from this group and can only rejoin if the owner invites you",
			})
			return
		}

		userIN, err := groups.UserIN(app.DB, req.GroupID, userID)
		if err != nil {
			helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
				"status":  false,
				"message": "could not check group membership",
			})
			return
		}

		if userIN {
			helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
				"status":  false,
				"message": "you already requested, were invited, or are a member of this group",
			})
			return
		}
	}

	if err := groups.SendGroupRequest(
		app.DB,
		userID,
		req.GroupID,
		req.Code,
	); err != nil {
		log.Println(err)

		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not process group request",
		})
		return
	}

	if req.Code == -1 {
		if err := notifications.DeleteGroupJoinRequests(app.DB, ownerID, req.GroupID, userID); err != nil {
			log.Println(err)
		}
	} else {
		g, err := groups.GetGroupData(app.DB, req.GroupID)
		if err != nil {
			log.Println(err)
		}

		actor := userID
		gid := req.GroupID

		delivered := app.notify(userID, models.NewNotification{
			UserID:          ownerID,
			Message:         fmt.Sprintf("%s requested to join the group \"%s\"", app.actorName(userID), g.Title),
			GroupJoinUserID: &actor,
			GroupID:         &gid,
		})

		if !delivered {
			if err := groups.SendGroupRequest(app.DB, userID, req.GroupID, -1); err != nil {
				log.Println("could not roll back join request:", err)
			}

			helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
				"status":  false,
				"message": "could not send join request",
			})
			return
		}
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status": true,
	})
}

func (app *App) GetGroupRequests(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	groupID, err := strconv.Atoi(r.URL.Query().Get("groupID"))
	if err != nil || groupID <= 0 {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid group",
		})
		return
	}

	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil || offset < 0 {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid offset",
		})
		return
	}

	ownerID, err := groups.GetGroupOwner(app.DB, groupID)
	if err != nil {
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not check group owner",
		})
		return
	}

	if ownerID != userID {
		helpers.WriteJson(w, http.StatusForbidden, map[string]any{
			"status":  false,
			"message": "only the group owner can view requests",
		})
		return
	}

	requests, err := groups.GetGroupRequests(app.DB, groupID, offset)
	if err != nil {
		log.Println(err)

		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not get group requests",
		})
		return
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status": true,
		"data":   requests,
	})
}

func (app *App) HandleGroupRequest(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	type Request struct {
		GroupID int `json:"groupID"`
		UserID  int `json:"userID"`
		Code    int `json:"code"`
	}

	var req Request

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid data",
		})
		return
	}

	if req.GroupID <= 0 || req.UserID <= 0 {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid request",
		})
		return
	}

	if req.Code != 1 && req.Code != -1 {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid request code",
		})
		return
	}

	ownerID, err := groups.GetGroupOwner(app.DB, req.GroupID)
	if err != nil {
		log.Println(err)
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not check group owner",
		})
		return
	}

	if ownerID != userID {
		helpers.WriteJson(w, http.StatusForbidden, map[string]any{
			"status":  false,
			"message": "only the group owner can handle requests",
		})
		return
	}

	pending, err := groups.HasJoinRequest(app.DB, req.GroupID, req.UserID)
	if err != nil {
		log.Println(err)
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not check request",
		})
		return
	}

	if !pending {
		helpers.WriteJson(w, http.StatusNotFound, map[string]any{
			"status":  false,
			"message": "request not found",
		})
		return
	}

	err = groups.HandleGroupRequest(
		app.DB,
		req.GroupID,
		req.UserID,
		req.Code,
	)

	if err != nil {
		log.Println(err)

		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not handle group request",
		})
		return
	}

	if err := notifications.DeleteGroupJoinRequests(app.DB, ownerID, req.GroupID, req.UserID); err != nil {
		log.Println(err)
	}

	if req.Code == 1 {
		app.notifyGroupRequestAccepted(ownerID, req.GroupID, req.UserID)
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status": true,
	})
}

func (app *App) CheckMessageAbility(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	targetID, err := strconv.Atoi(r.URL.Query().Get("targetID"))
	if err != nil {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid user ID",
		})
		return
	}

	if targetID <= 0 || targetID == userID {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid user ID",
		})
		return
	}

	canMessage, err := chats.CanSendMessage(app.DB, userID, targetID)
	if err != nil {
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not get user data",
		})
		return
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status":     true,
		"canMessage": canMessage,
	})
}

func (app *App) GetChatSuggestions(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	suggestions, err := chats.GetChatSuggestions(app.DB, userID, 5)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not get suggestions",
		})
		return
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status": true,
		"data":   suggestions,
	})
}

func (app *App) MarkChatRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	var body struct {
		GroupID int `json:"groupID"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.GroupID <= 0 {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid groupID",
		})
		return
	}

	inGroup, err := chats.UserInGroup(app.DB, userID, body.GroupID)
	if err != nil {
		log.Println(err)
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not verify group",
		})
		return
	}

	if !inGroup {
		helpers.WriteJson(w, http.StatusForbidden, map[string]any{
			"status":  false,
			"message": "you are not a member of this chat",
		})
		return
	}

	if err := chats.MarkGroupRead(app.DB, userID, body.GroupID); err != nil {
		log.Println(err)
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not mark chat as read",
		})
		return
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status": true,
	})
}
