package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"social/database/groups"
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

	userIN, err := groups.IsMember(app.DB, req.GroupID, userID)
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

	requested := 0

	for _, id := range req.UserIDs {
		result, err := app.addOrInvite(userID, id, req.GroupID, g.Title)

		if err != nil {
			log.Println(err)
			continue
		}

		switch result {
		case memberRequested:
			requested++
		}
	}

	if requested == 0 {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "you can only invite friends, followers or people you follow, and only if their settings allow it",
		})
		return
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status":  true,
		"message":   "members updated",
		"requested": requested,
		"sent":      requested,
	})
}
