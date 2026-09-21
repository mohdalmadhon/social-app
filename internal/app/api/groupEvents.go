package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"social/database/groups"
	"social/internal/helpers"
	"social/internal/models"
	"social/internal/validation"
)

func (app *App) AddGroupEvent(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)

	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	var event models.NewGroupEvent

	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid event",
		})
		return
	}

	event.Title = strings.TrimSpace(event.Title)
	event.Description = strings.TrimSpace(event.Description)
	event.EventTime = strings.TrimSpace(event.EventTime)

	if err := validation.ValidateGroupEvent(event); err != nil {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": err.Error(),
		})
		return
	}

	userIN, err := groups.UserIN(app.DB, event.GroupID, userID)

	if err != nil {
		log.Println(err)

		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not create event",
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

	eventID, err := groups.InsertGroupEvent(app.DB, event, userID)

	if err == sql.ErrNoRows {
		helpers.WriteJson(w, http.StatusForbidden, map[string]any{
			"status":  false,
			"message": "you are not a member of this group",
		})
		return
	}

	if err != nil {
		log.Println(err)

		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not create event",
		})
		return
	}

	created, err := groups.GetGroupEvent(app.DB, eventID, userID)

	if err != nil {
		log.Println(err)

		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not create event",
		})
		return
	}

	helpers.WriteJson(w, http.StatusCreated, map[string]any{
		"status":  true,
		"message": "event created",
		"data":    created,
	})
}

func (app *App) GetGroupEvents(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)

	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	groupID, err := strconv.Atoi(r.URL.Query().Get("groupID"))

	if err != nil || groupID <= 0 {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid group",
		})
		return
	}

	offset := 0

	if value := r.URL.Query().Get("offset"); value != "" {
		offset, err = strconv.Atoi(value)

		if err != nil || offset < 0 {
			helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
				"status":  false,
				"message": "invalid offset",
			})
			return
		}
	}

	limit := 10

	if value := r.URL.Query().Get("limit"); value != "" {
		parsed, err := strconv.Atoi(value)

		if err == nil && parsed > 0 && parsed <= 50 {
			limit = parsed
		}
	}

	userIN, err := groups.UserIN(app.DB, groupID, userID)

	if err != nil {
		log.Println(err)

		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not get events",
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

	events, err := groups.GetGroupEvents(app.DB, groupID, userID, limit, offset)

	if err != nil {
		log.Println(err)

		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not get events",
		})
		return
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status": true,
		"userId": userID,
		"data":   events,
	})
}

func (app *App) RespondGroupEvent(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)

	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	var response models.GroupEventResponse

	if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid response",
		})
		return
	}

	if err := validation.ValidateGroupEventResponse(response); err != nil {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": err.Error(),
		})
		return
	}

	groupID, err := groups.EventGroupID(app.DB, response.EventID)

	if err == sql.ErrNoRows {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "event does not exist",
		})
		return
	}

	if err != nil {
		log.Println(err)

		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not save response",
		})
		return
	}

	userIN, err := groups.UserIN(app.DB, groupID, userID)

	if err != nil {
		log.Println(err)

		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not save response",
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

	if err := groups.SetGroupEventResponse(app.DB, response.EventID, userID, response.Response); err != nil {
		log.Println(err)

		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not save response",
		})
		return
	}

	event, err := groups.GetGroupEvent(app.DB, response.EventID, userID)

	if err != nil {
		log.Println(err)

		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not save response",
		})
		return
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status":  true,
		"message": "response saved",
		"data":    event,
	})
}

func (app *App) GetGroupEventVotes(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)

	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	eventID, err := strconv.Atoi(r.URL.Query().Get("eventID"))

	if err != nil || eventID <= 0 {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid event",
		})
		return
	}

	response, err := strconv.Atoi(r.URL.Query().Get("response"))

	if err != nil || (response != 0 && response != 1) {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid response",
		})
		return
	}

	offset := 0

	if value := r.URL.Query().Get("offset"); value != "" {
		offset, err = strconv.Atoi(value)

		if err != nil || offset < 0 {
			helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
				"status":  false,
				"message": "invalid offset",
			})
			return
		}
	}

	limit := 10

	if value := r.URL.Query().Get("limit"); value != "" {
		parsed, err := strconv.Atoi(value)

		if err == nil && parsed > 0 && parsed <= 50 {
			limit = parsed
		}
	}

	groupID, err := groups.EventGroupID(app.DB, eventID)

	if err == sql.ErrNoRows {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "event does not exist",
		})
		return
	}

	if err != nil {
		log.Println(err)

		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not get votes",
		})
		return
	}

	userIN, err := groups.UserIN(app.DB, groupID, userID)

	if err != nil {
		log.Println(err)

		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not get votes",
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

	voters, err := groups.GetGroupEventVoters(app.DB, eventID, response, limit, offset)

	if err != nil {
		log.Println(err)

		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not get votes",
		})
		return
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status": true,
		"data":   voters,
	})
}
