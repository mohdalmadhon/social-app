package api

import (
	"database/sql"
	"encoding/json"
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

	userIN, err := groups.UserIN(app.DB, req.GroupID, userID)
	if err != nil {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "could not invite user",
		})
		return
	}

	if !userIN {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "could not invite user",
		})
		return
	}

	for _, id := range req.UserIDs {
		alreadyIN, err := groups.UserIN(app.DB, req.GroupID, id)
		if err != nil {
			helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
				"status":  false,
				"message": "could not invite user",
			})
			return
		}

		if alreadyIN {
			helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
				"status":  false,
				"message": "could not invite user",
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
			helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
				"status":  false,
				"message": "could not get group data",
			})
			return
		}

		g.ID = req.GroupID
		g.UserID = userID
		err = groups.SendInvites(app.DB, id, g)
		if err != nil {
			helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
				"status":  false,
				"message": "could not send invite",
			})
			return
		}
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status":  true,
		"message": "invite sent",
	})
	return
}
