package api

import (
	"log"
	"net/http"
	"slices"
	"social/database/preferences"
	"social/internal/helpers"
)

var ALLOWED_CHAT_PREFERENCES = []string{
	"any",
	"following",
	"following-followers",
	"friends",
	"friends-following",
	"none",
}

var ALLOWED_VISIBILITY_PREFERENCES = []string{
	"any",
	"none",
}

var ALLOWED_ADDITIONAL_INFO_PREFERENCES = []string{
	"any",
	"followers",
	"friends",
}

func (app *App) GetPreferences(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	prefs, err := preferences.GetPreferences(app.DB, userID)
	if err != nil {
		log.Println(err)
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not get preferences",
		})
		return
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status": true,
		"data":   prefs,
	})
}

func (app *App) ChangePerferance(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	Type := r.URL.Query().Get("type")
	value := r.URL.Query().Get("value")

	var allowed []string

	switch Type {
	case "chat":
		allowed = ALLOWED_CHAT_PREFERENCES
	case "email", "dob":
		allowed = ALLOWED_VISIBILITY_PREFERENCES
	case "additional":
		allowed = ALLOWED_ADDITIONAL_INFO_PREFERENCES
	default:
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid data",
		})
		return
	}

	if !slices.Contains(allowed, value) {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid data",
		})
		return
	}

	if err := preferences.ChangePreference(app.DB, Type, value, userID); err != nil {
		log.Println(err)
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not change data",
		})
		return
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status":  true,
		"message": "preference updated",
	})
}
