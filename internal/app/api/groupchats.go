package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"social/database/chats"
	"social/database/groups"
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

	log.Println(req.UserID)
	groupID := req.GroupID

	if groupID <= 0 {
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
			log.Println("here", exists, userIN)
			helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
				"status":  false,
				"message": "could not send message",
			})
			return
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
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid groupID",
		})
		return
	}

	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil {
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

	if !exists {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid group",
		})
		return
	}

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
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid request",
		})
		return
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

	log.Println(userIDs)
	for _, id := range ids {
		if id == userID {
			continue
		}
		isFriend, err := users.IsFriend(app.DB, userID, id)
		if err != nil {
			log.Println(err, "here")
			helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
				"status":  false,
				"message": "failed to check data",
			})
			return
		}

		if isFriend {
			err = groups.AddMembers(app.DB, g.ID, userID)
		} else {
			err = groups.SendInvites(app.DB, id, g)
			if err != nil {
				log.Println(err, "err1")
				continue
			}
			requests = append(requests, id)
		}

		if err != nil {
			log.Println(err, "here2")
			helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
				"status":  false,
				"message": "failed to check data",
			})
			return
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

	if err := groups.ChangeStatus(app.DB, userID, req.Status, req.GroupID); err != nil {
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not update group status",
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

	if req.Code == 0 {
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
				"message": "you are already a member of this group",
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

	member, err := groups.UserIN(app.DB, groupID, userID)
	if err != nil {
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not check group membership",
		})
		return
	}

	if !member {
		helpers.WriteJson(w, http.StatusForbidden, map[string]any{
			"status":  false,
			"message": "you are not a member of this group",
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
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not check group membership",
		})
		return
	}

	if ownerID != userID {
		if err != nil {
			helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
				"status":  false,
				"message": "could not check group membership",
			})
			return
		}
	}

	member, err := groups.UserIN(app.DB, req.GroupID, userID)
	if err != nil {
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not check group membership",
		})
		return
	}

	if !member {
		helpers.WriteJson(w, http.StatusForbidden, map[string]any{
			"status":  false,
			"message": "you are not a member of this group",
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

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status": true,
	})
}
