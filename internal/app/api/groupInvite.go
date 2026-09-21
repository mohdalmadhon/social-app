package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"social/database/groups"
	"social/database/users"
	"social/internal/helpers"
)

func (app *App) InviteMember(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	type Request struct {
		GroupID int   `json:"groupID"`
		UserIDs []int `json:"userIDs"`
	}

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid data",
		})
		return
	}

	if req.GroupID <= 0 || len(req.UserIDs) == 0 {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid group or users",
		})
		return
	}

	userIN, err := groups.UserIN(app.DB, req.GroupID, userID)
	if err != nil {
		log.Println(err)
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not check group membership",
		})
		return
	}

	if !userIN {
		helpers.WriteJson(w, http.StatusForbidden, map[string]any{
			"status":  false,
			"message": "you are not a member of this group",
		})
		return
	}

	g, err := groups.GetGroupData(app.DB, req.GroupID)
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

	g.ID = req.GroupID
	g.UserID = userID

	sent := 0

	for _, id := range req.UserIDs {
		if id == userID {
			continue
		}

		alreadyIN, err := groups.UserIN(app.DB, req.GroupID, id)
		if err != nil {
			log.Println(err)
			continue
		}

		if alreadyIN {
			log.Println("already in")
			continue
		}

		isFriend, err := users.IsFriend(app.DB, userID, id)
		if err != nil {
			continue
		}

		if isFriend {
			err = groups.AddMembers(app.DB, req.GroupID, id)
			if err != nil {
				continue
			}
		} else {
			err = groups.SendInvites(app.DB, id, g)
			if err != nil {
				log.Println(err)
				continue
			}
		}
		sent++
	}

	if sent == 0 {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "no invites were sent",
		})
		return
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status":  true,
		"message": "invite sent",
		"sent":    sent,
	})
}
