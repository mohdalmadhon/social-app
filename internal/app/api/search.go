package api

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"social/database/groups"
	"social/database/posts"
	"social/database/users"
	"social/internal/helpers"
	"social/internal/models"
)

const (
	searchMaxLength     = 100
	searchPageSize      = 12
	searchPostsPageSize = 10
)

func (app *App) GlobalSearchGroups(w http.ResponseWriter, r *http.Request) {
	serveSearch(w, r, searchPageSize, func(userID int, search string, limit, offset int) ([]models.SearchGroup, error) {
		return groups.SearchGroups(app.DB, userID, search, limit, offset)
	})
}

func (app *App) GlobalSearchUsers(w http.ResponseWriter, r *http.Request) {
	serveSearch(w, r, searchPageSize, func(userID int, search string, limit, offset int) ([]models.SearchUser, error) {
		return users.SearchUsers(app.DB, userID, search, limit, offset)
	})
}

func (app *App) GlobalSearchPosts(w http.ResponseWriter, r *http.Request) {
	serveSearch(w, r, searchPostsPageSize, func(userID int, search string, limit, offset int) ([]models.Post, error) {
		return posts.SearchPosts(app.DB, userID, search, limit, offset)
	})
}

// serveSearch validates the shared search parameters (search, offset), runs
// the given query for one page of results and writes the JSON response.
// fetch is asked for one row more than the page size so that "hasMore" can
// be answered without a second query.
func serveSearch[T any](
	w http.ResponseWriter,
	r *http.Request,
	pageSize int,
	fetch func(userID int, search string, limit, offset int) ([]T, error),
) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	search := strings.TrimSpace(r.URL.Query().Get("search"))

	if utf8.RuneCountInString(search) > searchMaxLength {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "search is too long",
		})
		return
	}

	offset := 0

	if value := r.URL.Query().Get("offset"); value != "" {
		var err error

		offset, err = strconv.Atoi(value)

		if err != nil || offset < 0 {
			helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
				"status":  false,
				"message": "invalid offset",
			})
			return
		}
	}

	if search == "" {
		helpers.WriteJson(w, http.StatusOK, map[string]any{
			"status":  true,
			"data":    []T{},
			"hasMore": false,
		})
		return
	}

	results, err := fetch(userID, search, pageSize+1, offset)

	if err != nil {
		log.Println(err)
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not search",
		})
		return
	}

	hasMore := len(results) > pageSize

	if hasMore {
		results = results[:pageSize]
	}

	if results == nil {
		results = []T{}
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status":  true,
		"data":    results,
		"hasMore": hasMore,
	})
}

func (app *App) GetFollowers_Following(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "could not authorize user",
		})
		return
	}

	search := r.URL.Query().Get("search")
	log.Println(search)
	users, err := users.Get_followers_following_chatList(app.DB, userID, search)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not get users",
		})
		return
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status": true,
		"data":   users,
	})
}
