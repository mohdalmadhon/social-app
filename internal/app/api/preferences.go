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

	switch Type {
	case "chat":
		if !slices.Contains(ALLOWED_CHAT_PREFERENCES, value) {
			helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
				"status":  false,
				"message": "invalid data",
			})
			return
		}
		err := preferences.ChangeChatPreferences(app.DB, value, userID)
		if err != nil {
			log.Println(err)
			helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
				"status":  false,
				"message": "could not change data",
			})
			return
		}
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status":  true,
		"message": "chat perferance updated",
	})
	return
}
