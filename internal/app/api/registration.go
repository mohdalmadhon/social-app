package api

import (
	"database/sql"
	"log"
	"net/http"
	"social/database/users"
	"social/internal/helpers"
	"social/internal/models"
	"social/internal/validation"
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

type App struct {
	DB *sql.DB
	H  *Hub
}

type Hub struct {
	Conn map[int]*websocket.Conn
	Mu   sync.RWMutex
}

func (app *App) RegisterUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20)

	if err != nil {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "Bad request, form data is too big",
		})
		return
	}

	dobValue := r.FormValue("dob")

	dob, err := time.Parse("2006-01-02", dobValue)
	if err != nil {
		log.Println(err)

		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "Invalid date of birth",
		})
		return
	}

	userData := models.UserRegistration{
		FirstName: r.FormValue("FirstName"),
		LastName:  r.FormValue("LastName"),
		UserName:  r.FormValue("UserName"),
		Email:     r.FormValue("Email"),
		Password:  r.FormValue("Password"),
		About:     r.FormValue("About"),
		DOB:       dob,
		Avatar:    "",
	}

	err = validation.ValidateRegisterData(&userData)
	if err != nil {
		log.Println(err)

		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": err.Error(),
		})
		return
	}

	file, header, err := r.FormFile("Avatar")
	if err != nil {
		if err != http.ErrMissingFile {
			helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
				"status":  false,
				"message": "invalid avatar upload",
			})
			return
		}
	} else {
		defer file.Close()

		avatarPath, err := helpers.SaveUploads(file, header, "avatar")
		if err != nil {
			log.Println(err)

			helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
				"status":  false,
				"message": "could not save avatar",
			})
			return
		}

		userData.Avatar = avatarPath
	}

	hashedPassword, err := helpers.HashPassword(userData.Password)
	if err != nil {
		log.Println(err)

		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not hash password",
		})
		return
	}

	userData.Password = hashedPassword

	if err := users.RegisterUser(app.DB, &userData); err != nil {

		status, message := helpers.NormalizeSQLError(err)

		helpers.WriteJson(w, status, map[string]any{
			"status":  false,
			"message": message,
		})
		return
	}

	helpers.WriteJson(w, http.StatusCreated, map[string]any{
		"status":  true,
		"message": "Registration successful",
	})
}
