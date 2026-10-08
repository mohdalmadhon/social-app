package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"social/database/groups"
	"social/database/users"
	"social/internal/helpers"
	"social/internal/models"
	"social/internal/validation"
	"strings"
)

func (app *App) GetPostGroups(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	groups, err := groups.GetGroups(app.DB, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			helpers.WriteJson(w, http.StatusOK, map[string]any{
				"status": true,
				"data":   nil,
			})
			return
		}
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
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

func (app *App) AddPostGroup(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	var newGroup models.NewGroup
	newGroup.UserID = userID
	r.Body = http.MaxBytesReader(w, r.Body, validation.MaxJSONBody)

	err := json.NewDecoder(r.Body).Decode(&newGroup)
	if err != nil {
		log.Println(err)
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid users",
		})
		return
	}

	if err := validation.ValidateNewGroup(&newGroup); err != nil {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": err.Error(),
		})
		return
	}

	valid := func() bool {
		for _, id := range newGroup.Users {
			err := users.UserExists(app.DB, id)
			if err != nil {
				return false
			}
		}
		return true
	}

	if !valid() {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "user does not exists",
		})
		return
	}

	err = groups.AddGroup(app.DB, newGroup)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			log.Println(err)
			helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
				"status":  false,
				"message": "error already exists",
			})
			return
		}
		log.Println(err)
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not add group",
		})
		return
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status":  true,
		"message": "group added",
	})
}

func (app *App) DeletePostGroup(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	var data struct {
		GroupID int
	}
	r.Body = http.MaxBytesReader(w, r.Body, validation.MaxJSONBody)

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		log.Println(err)
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid group id",
		})
		return
	}

	if err := groups.DeleteGroup(app.DB, data.GroupID, userID); err != nil {
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not delete group",
		})
		return
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status":  true,
		"message": "group deleted",
	})
}

func (app *App) UpdatePostGroup(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	var group models.NewGroup
	r.Body = http.MaxBytesReader(w, r.Body, validation.MaxJSONBody)

	if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
		log.Println(err)
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid group data",
		})
		return
	}

	if group.GroupID <= 0 {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "missing group id",
		})
		return
	}

	if err := validation.ValidateNewGroup(&group); err != nil {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": err.Error(),
		})
		return
	}

	group.UserID = userID
	if err := groups.UpdateGroup(app.DB, group); err != nil {
		log.Println(err)
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "failed to update group",
		})
		return
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status":  true,
		"message": "group updated",
	})
}
