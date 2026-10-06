package api

import (
	"encoding/json"
	"log"

	"social/database/chats"
)

func (app *App) broadcastGroupTyping(userID, groupID int, typing bool) {
	memberIDs, err := chats.GetGroupMembersIds(app.DB, groupID)
	if err != nil {
		log.Println("could not get group members for typing:", err)
		return
	}

	payload, err := json.Marshal(map[string]any{
		"type": "typing",
		"data": map[string]any{
			"userID":  userID,
			"groupID": groupID,
			"typing":  typing,
			"isGroup": true,
		},
	})

	if err != nil {
		log.Println("marshal group typing error:", err)
		return
	}

	for _, memberID := range memberIDs {
		if memberID == userID {
			continue
		}

		app.H.Mu.RLock()
		conn, ok := app.H.Conn[memberID]
		app.H.Mu.RUnlock()

		if !ok {
			continue
		}

		if _, err := conn.Write(payload); err != nil {
			log.Println("websocket group typing write error:", err)
		}
	}
}
