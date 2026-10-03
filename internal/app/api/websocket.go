package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"social/database/chats"
	"social/database/posts"
	"social/database/preferences"
	"social/database/users"
	"social/internal/helpers"
	"social/internal/models"
	"strings"

	"golang.org/x/net/websocket"
)

func (app *App) HandleWS(ws *websocket.Conn) {
	userID, ok := ws.Request().Context().Value("userID").(int)
	if !ok {
		log.Println("invalid user ID")
		return
	}

	app.register(userID, ws)
	defer app.unregister(userID)

	app.readLoop(userID, ws)
}

func (app *App) readLoop(userID int, ws *websocket.Conn) {
	for {
		var rawMessage string

		err := websocket.Message.Receive(ws, &rawMessage)
		if err != nil {
			if err == io.EOF {
				break
			}

			log.Println("websocket read error:", err)
			break
		}

		var payload models.WSPayload

		if err := json.Unmarshal([]byte(rawMessage), &payload); err != nil {
			log.Println("invalid websocket payload:", err)
			continue
		}

		switch payload.Type {
		case "privateMessage":
			app.handleMessage(userID, payload.Data, "message")

		case "notification":
			log.Println("notification received")

		case "privateMessage/invite":
			app.handleInvite(userID, payload.Data)

		case "postGroup":
			app.handleMessage(userID, payload.Data, "postGroup")
		case "post-message":

		default:
			log.Println("unknown websocket type:", payload.Type)
		}
	}
}

func (app *App) sendNotification(userID int, targetID int, data models.NewNotification, notificationID int, unread int) {
	userData, err := users.GetUserSimpleData(app.DB, userID)

	if err != nil {
		errMsg := map[string]any{
			"type":    "notification",
			"error":   true,
			"message": "could not send notification",
		}

		payload, err := json.Marshal(errMsg)
		if err != nil {
			log.Println(err)
			return
		}

		app.H.Mu.Lock()
		conn := app.H.Conn[userID]
		app.H.Mu.Unlock()

		if conn != nil {
			if _, err := conn.Write(payload); err != nil {
				log.Println(err)
			}
		}

		return
	}

	msgWord := helpers.GetNotificationType(data)

	name := userData.UserName

	if name == "" {
		name = strings.TrimSpace(userData.FirstName + " " + userData.LastName)
	}

	message := data.Message

	if message == "" {
		message = fmt.Sprintf("new %s from %s", msgWord, name)
	}

	kind := msgWord

	if data.FollowRequestUserID != nil {
		kind = "follow_request"
	}

	msg := map[string]any{
		"type":    "notification",
		"error":   false,
		"message": message,
		"id":      notificationID,
		"kind":    kind,
		"actor": map[string]any{
			"id":         userData.ID,
			"firstName":  userData.FirstName,
			"lastName":   userData.LastName,
			"avatarPath": userData.Avatar,
		},
	}

	if unread >= 0 {
		msg["unread"] = unread
	}

	if data.PostIDTag != nil {
		msg["post_id"] = *data.PostIDTag

		if imagePath, err := posts.GetPostImage(app.DB, *data.PostIDTag); err == nil && imagePath != "" {
			msg["image_path"] = imagePath
		}
	}

	msgPayload, err := json.Marshal(msg)
	if err != nil {
		log.Println(err)
		return
	}

	app.H.Mu.Lock()
	conn := app.H.Conn[targetID]
	app.H.Mu.Unlock()

	if conn != nil {
		if _, err := conn.Write(msgPayload); err != nil {
			log.Println(err)
		}
	}
}

func (app *App) handleMessage(userID int, data json.RawMessage, Type string) {
	var msg models.IncomingMessage

	if Type == "postGroup" {
		var postMessage models.PostMessage

		if err := json.Unmarshal(data, &postMessage); err != nil {
			log.Println("invalid post message payload:", err)
			return
		}

		if postMessage.GroupID <= 0 {
			log.Println("invalid group ID:", postMessage.GroupID)
			return
		}

		content, err := json.Marshal(postMessage)
		if err != nil {
			log.Println("marshal post message error:", err)
			return
		}

		groupID := postMessage.GroupID

		if err := chats.AddMessages(
			app.DB,
			string(content),
			userID,
			groupID,
		); err != nil {
			log.Println("add post message error:", err)
			return
		}

		sender, err := users.GetUserSimpleData(app.DB, userID)
		if err != nil {
			log.Println("get sender error:", err)
			return
		}

		message := models.Message{
			Content: string(content),
			Sender: models.UserRegistration{
				ID:        userID,
				FirstName: sender.FirstName,
				LastName:  sender.LastName,
				Avatar:    sender.Avatar,
			},
			GroupID: groupID,
		}

		app.sendToUsers(message, groupID, userID)

		return
	}

	if err := json.Unmarshal(data, &msg); err != nil {
		log.Println("invalid message payload:", err)
		return
	}

	if msg.Content == "" {
		return
	}

	groupID := msg.GroupID

	if groupID <= 0 {
		if msg.UserID <= 0 || msg.UserID == userID {
			app.sendMessageError(userID, msg.ClientID, "invalid user")
			return
		}

		canMessage, err := chats.CanSendMessage(app.DB, userID, msg.UserID)
		if err != nil {
			log.Println("message permission check error:", err)
			app.sendMessageError(userID, msg.ClientID, "could not verify message permission")
			return
		}

		if !canMessage {
			app.sendMessageError(userID, msg.ClientID, "could not send message because of user preference")
			return
		}

		existingGroupID, err := chats.HasPrivateChat(
			app.DB,
			userID,
			msg.UserID,
		)

		if err != nil {
			log.Println("private chat lookup error:", err)
			app.sendMessageError(userID, msg.ClientID, "could not verify chat")
			return
		}

		if existingGroupID != -1 {
			groupID = existingGroupID
		} else {
			groupID, err = chats.MakePrivateChat(
				app.DB,
				userID,
				msg.UserID,
			)

			if err != nil {
				log.Println("private chat creation error:", err)
				app.sendMessageError(userID, msg.ClientID, "could not create chat")
				return
			}
		}
	} else {
		isPrivate, targetID, err := chats.IsPrivateChat(app.DB, groupID, userID)
		if err != nil {
			log.Println("chat lookup error:", err)
			app.sendMessageError(userID, msg.ClientID, "could not verify chat")
			return
		}

		if isPrivate {
			canMessage, err := chats.CanSendMessage(app.DB, userID, targetID)
			if err != nil {
				log.Println("message permission check error:", err)
				app.sendMessageError(userID, msg.ClientID, "could not verify message permission")
				return
			}

			if !canMessage {
				app.sendMessageError(userID, msg.ClientID, "could not send message because of user preference")
				return
			}
		}
	}

	if err := chats.AddMessages(
		app.DB,
		msg.Content,
		userID,
		groupID,
	); err != nil {
		log.Println("add message error:", err)
		app.sendMessageError(userID, msg.ClientID, "could not send message")
		return
	}

	sender, err := users.GetUserSimpleData(app.DB, userID)
	if err != nil {
		log.Println("get sender error:", err)
		return
	}

	message := models.Message{
		Content: msg.Content,
		Sender: models.UserRegistration{
			ID:        userID,
			FirstName: sender.FirstName,
			LastName:  sender.LastName,
			Avatar:    sender.Avatar,
		},
		GroupID: groupID,
	}

	app.sendToUsers(message, groupID, userID)
}

func (app *App) handleInvite(userID int, data json.RawMessage) {
	var invite models.GroupInvite

	if err := json.Unmarshal(data, &invite); err != nil {
		log.Println("invalid invite payload:", err)
		return
	}

	if invite.GroupData.ID <= 0 {
		log.Println("invalid group ID")
		return
	}

	if len(invite.Users) == 0 {
		log.Println("no users in invite")
		return
	}

	sender, err := users.GetUserSimpleData(app.DB, userID)
	if err != nil {
		log.Println("get sender error:", err)
		return
	}

	inviteContent := map[string]any{
		"type": "invite",
		"group": map[string]any{
			"id":     invite.GroupData.ID,
			"name":   invite.GroupData.Name,
			"avatar": invite.GroupData.Avatar,
		},
		"user": map[string]any{
			"id":        sender.ID,
			"firstName": sender.FirstName,
			"lastName":  sender.LastName,
			"avatar":    sender.Avatar,
		},
	}

	content, err := json.Marshal(inviteContent)
	if err != nil {
		log.Println("marshal invite error:", err)
		return
	}

	for _, invitedUserID := range invite.Users {
		if invitedUserID <= 0 || invitedUserID == userID {
			continue
		}
		_, err = app.DB.Exec(`INSERT INTO groups_users (group_id, user_id, status) VALUES (?,?,0)`, invite.GroupData.ID, invitedUserID)
		if err != nil {
			log.Println(err)
			continue
		}

		privateChatID, err := chats.HasPrivateChat(
			app.DB,
			userID,
			invitedUserID,
		)

		if err != nil {
			log.Println("private chat lookup error:", err)
			continue
		}

		if privateChatID == -1 {
			privateChatID, err = chats.MakePrivateChat(
				app.DB,
				userID,
				invitedUserID,
			)

			if err != nil {
				log.Println("private chat creation error:", err)
				continue
			}
		}

		err = chats.AddMessages(
			app.DB,
			string(content),
			userID,
			privateChatID,
		)

		if err != nil {
			log.Println("add invite message error:", err)
			continue
		}

		message := models.Message{
			Content: string(content),
			Sender: models.UserRegistration{
				ID:        sender.ID,
				FirstName: sender.FirstName,
				LastName:  sender.LastName,
				Avatar:    sender.Avatar,
			},
			GroupID: privateChatID,
		}

		app.sendToUsers(
			message,
			privateChatID,
			userID,
		)
	}
}

func (app *App) sendToUsers(
	msg models.Message,
	groupID int,
	userID int,
) {

	log.Println(groupID)
	ids, err := chats.GetGroupMembersIds(
		app.DB,
		groupID,
	)

	if err != nil {
		log.Println("get group members error:", err)
		return
	}

	isPrivate, groupName, err := chats.GetChatMeta(app.DB, groupID)

	if err != nil {
		log.Println("get chat meta error:", err)
		return
	}

	buildResponse := func(silent bool) ([]byte, error) {
		return json.Marshal(map[string]any{
			"type":      "message",
			"data":      msg,
			"isPrivate": isPrivate,
			"groupName": groupName,
			"silent":    silent,
		})
	}

	response, err := buildResponse(false)

	if err != nil {
		log.Println("marshal websocket response error:", err)
		return
	}

	silentResponse, err := buildResponse(true)

	if err != nil {
		log.Println("marshal websocket response error:", err)
		return
	}

	for _, id := range ids {
		if id == userID {
			continue
		}

		payload := response

		allowed, err := preferences.ShouldNotify(app.DB, id, userID, "message")

		if err != nil {
			log.Println("failed to check message notification preference:", err)
		} else if !allowed {
			payload = silentResponse
		}

		app.H.Mu.RLock()
		client, ok := app.H.Conn[id]
		app.H.Mu.RUnlock()

		if !ok {
			continue
		}

		if _, err := client.Write(payload); err != nil {
			log.Println("websocket write error:", err)
		}
	}
}

func (app *App) register(userID int, ws *websocket.Conn) {
	app.H.Mu.Lock()
	defer app.H.Mu.Unlock()

	app.H.Conn[userID] = ws
}

func (app *App) unregister(userID int) {
	app.H.Mu.Lock()
	defer app.H.Mu.Unlock()

	delete(app.H.Conn, userID)
}

func (app *App) sendMessageError(userID int, clientID string, message string) {
	response, err := json.Marshal(map[string]any{
		"type": "notification",
		"data": map[string]any{
			"error":    true,
			"clientID": clientID,
			"message":  message,
		},
	})

	if err != nil {
		log.Println("marshal notification error:", err)
		return
	}

	app.H.Mu.RLock()
	conn, ok := app.H.Conn[userID]
	app.H.Mu.RUnlock()

	if !ok {
		log.Println("user websocket connection not found:", userID)
		return
	}

	if _, err := conn.Write(response); err != nil {
		log.Println("websocket notification error:", err)
	}
}
