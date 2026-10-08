package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"social/database/groups"
	"social/internal/helpers"
	"social/internal/validation"
)

func (app *App) pushGroupRemoved(userID, groupID int, reason string) {
	payload, err := json.Marshal(map[string]any{
		"type":    "groupRemoved",
		"groupID": groupID,
		"reason":  reason,
	})

	if err != nil {
		log.Println(err)
		return
	}

	app.H.Mu.RLock()
	conn := app.H.Conn[userID]
	app.H.Mu.RUnlock()

	if conn == nil {
		return
	}

	if _, err := conn.Write(payload); err != nil {
		log.Println(err)
	}
}

func (app *App) KickMember(w http.ResponseWriter, r *http.Request) {
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
	}

	var req Request

	r.Body = http.MaxBytesReader(w, r.Body, validation.MaxJSONBody)

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
			"message": "invalid group or user",
		})
		return
	}

	exists, err := groups.GroupExists(app.DB, req.GroupID)
	if err != nil {
		log.Println(err)
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not check group",
		})
		return
	}

	if !exists {
		helpers.WriteJson(w, http.StatusNotFound, map[string]any{
			"status":  false,
			"message": "could not find group",
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
			"message": "only the group owner can remove members",
		})
		return
	}

	if req.UserID == userID {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "you cannot remove yourself from your own group",
		})
		return
	}

	if err := groups.KickMember(app.DB, req.GroupID, req.UserID, userID); err != nil {
		if err == sql.ErrNoRows {
			helpers.WriteJson(w, http.StatusNotFound, map[string]any{
				"status":  false,
				"message": "user is not a member of this group",
			})
			return
		}

		log.Println(err)
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not remove member",
		})
		return
	}

	app.pushGroupRemoved(req.UserID, req.GroupID, "kicked")

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status":  true,
		"message": "member removed",
	})
}

func (app *App) LeaveGroup(w http.ResponseWriter, r *http.Request) {
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
	}

	var req Request

	r.Body = http.MaxBytesReader(w, r.Body, validation.MaxJSONBody)

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

	exists, err := groups.GroupExists(app.DB, req.GroupID)
	if err != nil {
		log.Println(err)
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not check group",
		})
		return
	}

	if !exists {
		helpers.WriteJson(w, http.StatusNotFound, map[string]any{
			"status":  false,
			"message": "could not find group",
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

	if ownerID == userID {
		helpers.WriteJson(w, http.StatusForbidden, map[string]any{
			"status":  false,
			"message": "the group owner cannot leave the group",
		})
		return
	}

	if err := groups.LeaveGroup(app.DB, req.GroupID, userID); err != nil {
		if err == sql.ErrNoRows {
			helpers.WriteJson(w, http.StatusNotFound, map[string]any{
				"status":  false,
				"message": "you are not a member of this group",
			})
			return
		}

		log.Println(err)
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not leave group",
		})
		return
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status":  true,
		"message": "you left the group",
	})
}
