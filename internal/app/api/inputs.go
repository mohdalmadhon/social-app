package api

import (
	"net/http"

	"social/internal/helpers"
	"social/internal/validation"
)

func readSearch(w http.ResponseWriter, r *http.Request) (string, bool) {
	value, err := validation.ValidateSearch(r.URL.Query().Get("search"))

	if err != nil {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": err.Error(),
		})
		return "", false
	}

	return value, true
}

func readIDList(w http.ResponseWriter, label string, ids []int) bool {
	if err := validation.ValidateIDList(label, ids); err != nil {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": err.Error(),
		})
		return false
	}

	return true
}
