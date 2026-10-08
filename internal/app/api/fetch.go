package api

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"social/database/profiles"
	"social/database/users"
	"social/internal/helpers"
	"social/internal/validation"
	"strconv"
	"strings"
)

func (app *App) SearchLocation(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))

	if query == "" {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "no data provided",
		})
		return
	}

	if len([]rune(query)) > validation.MaxSearch {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "search text is too long",
		})
		return
	}

	if len([]rune(query)) < 2 {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "search text is too short",
		})
		return
	}

	if len([]rune(query)) > 150 {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "search text is too long",
		})
		return
	}

	lang := strings.TrimSpace(r.Header.Get("Accept-Language"))

	if lang == "" || len(lang) > 100 {
		lang = "en"
	}

	cacheKey := strings.ToLower(query) + "|" + lang

	if cached, ok := locationCacheGet(cacheKey); ok {
		w.Header().Set("Content-Type", "application/json")
		w.Write(cached)
		return
	}

	endpoint := "https://nominatim.openstreetmap.org/search" +
		"?format=jsonv2" +
		"&addressdetails=1" +
		"&dedupe=1" +
		"&limit=8" +
		"&accept-language=" + url.QueryEscape(lang) +
		"&q=" + url.QueryEscape(query)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		log.Println(err)

		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "failed to create location request",
		})
		return
	}

	req.Header.Set("User-Agent", "MySocialNetwork/1.0")

	locationThrottle()

	resp, err := locationClient.Do(req)
	if err != nil {
		log.Println(err)

		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "failed to search location",
		})
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("Nominatim returned status:", resp.StatusCode)

		helpers.WriteJson(w, http.StatusBadGateway, map[string]any{
			"status":  false,
			"message": "location API returned an error",
		})
		return
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		log.Println(err)

		helpers.WriteJson(w, http.StatusBadGateway, map[string]any{
			"status":  false,
			"message": "could not read location response",
		})
		return
	}

	locationCacheSet(cacheKey, body)

	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func (app *App) GetFriends(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	targetIDParam := r.URL.Query().Get("targetid")
	targetID := userID
	if targetIDParam != "" && targetIDParam != "null" && targetIDParam != "undefined" {
		parsedTargetID, err := strconv.Atoi(targetIDParam)
		if err != nil {
			helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
				"status":  false,
				"message": "invalid targetid",
			})
			return
		}
		targetID = parsedTargetID
	}

	searchValue, searchOK := readSearch(w, r)
	if !searchOK {
		return
	}
	queryOffset := r.URL.Query().Get("offset")
	offset, err := strconv.Atoi(queryOffset)

	if err != nil || offset < 0 {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid offset",
		})
		return
	}

	if searchValue == "" {
		friends, err := users.GetFriends(app.DB, targetID, offset)
		if err != nil {
			log.Println(err)
			helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
				"status":  false,
				"message": "could not get user friends",
			})
			return
		}
		helpers.WriteJson(w, http.StatusOK, map[string]any{
			"status":  true,
			"message": "friends fetched",
			"data":    friends,
		})
		return
	}

	friends, err := users.SearchFriends(app.DB, targetID, searchValue)
	if err != nil {
		log.Println(err)
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not get friends",
			"data":    friends,
		})
		return
	}
	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status":  true,
		"message": "friends fetched",
		"data":    friends,
	})
}

func (app *App) SearchFollows(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)

	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	targetID := userID
	queryID := r.URL.Query().Get("targetid")

	if queryID != "" && queryID != "null" && queryID != "undefined" {
		parsedID, err := strconv.Atoi(queryID)

		if err != nil || parsedID <= 0 {
			log.Println(err)
			helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
				"status":  false,
				"message": "invalid targetid",
			})
			return
		}

		targetID = parsedID
	}

	searchValue, searchOK := readSearch(w, r)
	if !searchOK {
		return
	}

	follows, err := profiles.SearchFollows(app.DB, targetID, searchValue)

	if err != nil {
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not find follows",
		})
		return
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status": true,
		"data":   follows,
	})
}

func (app *App) SearchFollowing(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)

	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	targetID := userID
	queryID := r.URL.Query().Get("targetid")

	if queryID != "" && queryID != "null" && queryID != "undefined" {
		parsedID, err := strconv.Atoi(queryID)

		if err != nil || parsedID <= 0 {
			log.Println(err)
			helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
				"status":  false,
				"message": "invalid targetid",
			})
			return
		}

		targetID = parsedID
	}

	searchValue, searchOK := readSearch(w, r)
	if !searchOK {
		return
	}

	following, err := profiles.SearchFollowing(app.DB, targetID, searchValue)

	if err != nil {
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not find following",
		})
		return
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status": true,
		"data":   following,
	})
}
