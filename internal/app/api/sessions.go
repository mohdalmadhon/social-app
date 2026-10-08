package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"social/database/users"
	"social/internal/app/tokens"
	"social/internal/helpers"
	"social/internal/models"
	"social/internal/validation"
	"time"
)

func (app *App) LoggingUser(w http.ResponseWriter, r *http.Request) {
	var logger models.UserLogger

	r.Body = http.MaxBytesReader(w, r.Body, 4<<10)

	if err := json.NewDecoder(r.Body).Decode(&logger); err != nil {
		log.Println(err)
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "could not get data",
		})
		return
	}

	identifier, err := validation.ValidateLogin(logger.Identifier, logger.Pass)
	if err != nil {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": err.Error(),
		})
		return
	}

	logger.Identifier = identifier

	hashedPassword, err := users.GetHashedPassowrd(app.DB, logger.Identifier)
	if err != nil {
		if err == sql.ErrNoRows {
			helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
				"status":  false,
				"message": "user identifier or password is incorrect",
			})
			return
		} else {
			helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
				"status":  false,
				"message": "could not get user data",
			})
			return
		}
	}

	if match := helpers.AuthonticateUser(logger.Pass, hashedPassword); !match {
		helpers.WriteJson(w, http.StatusUnauthorized, map[string]any{
			"status":  false,
			"message": "user identifier or password is incorrect",
		})
		return
	}

	userID := users.GetUserID(app.DB, logger.Identifier)
	if userID == -1 {
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not get user data",
		})
		return
	}

	token, err := tokens.GenerateToken(userID)
	if err != nil {
		log.Println(err)
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not create authentication token",
		})
		return
	}

	cookie := http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(24 * 30 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, &cookie)

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status":  true,
		"message": "logged in",
	})
}

func (app *App) AuthorizeSession(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value("userID").(int)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]any{
		"status":  true,
		"message": "valid session",
		"userID":  userID,
	})
}

func (app *App) DeleteSession(w http.ResponseWriter, r *http.Request) {

	http.SetCookie(w, &http.Cookie{
		Name:   "token",
		Value:  "",
		MaxAge: -1,
		Path:   "/",
	})

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status":  true,
		"message": "logged out",
	})
}
