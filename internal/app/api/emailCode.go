package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"social/database/users"
	"social/internal/app/otp"
	"social/internal/helpers"
	"social/internal/validation"
	"strings"
)

type emailCodeRequest struct {
	Email string `json:"email"`
}

type verifyCodeRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func isSixDigits(code string) bool {
	if len(code) != otp.CodeLength {
		return false
	}

	for _, r := range code {
		if r < '0' || r > '9' {
			return false
		}
	}

	return true
}

func (app *App) CheckAvailability(w http.ResponseWriter, r *http.Request) {
	kind := r.URL.Query().Get("type")
	value := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("value")))

	var field string

	switch kind {
	case "name":
		field = "username"
	case "email":
		field = "email"
	default:
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "invalid check type",
		})
		return
	}

	if value == "" {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "value is required",
		})
		return
	}

	available, err := users.IsAvailable(app.DB, field, value)
	if err != nil {
		log.Println(err)

		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not check availability",
		})
		return
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status":    true,
		"avilable":  available,
		"available": available,
	})
}

func (app *App) SendEmailCode(w http.ResponseWriter, r *http.Request) {
	var req emailCodeRequest

	r.Body = http.MaxBytesReader(w, r.Body, 1<<10)

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "could not read request",
		})
		return
	}

	email := normalizeEmail(req.Email)

	if err := validation.ValidateEmail(email); err != nil {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "Invalid email",
		})
		return
	}

	available, err := users.IsAvailable(app.DB, "email", email)
	if err != nil {
		log.Println(err)

		log.Println(err)
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "Something went wrong. Please try again.",
		})
		return
	}

	if !available {
		helpers.WriteJson(w, http.StatusConflict, map[string]any{
			"status":  false,
			"message": "This email is already registered.",
		})
		return
	}

	code, err := app.OTP.Issue(email)
	if err != nil {
		var limit *otp.RateLimitError

		if errors.As(err, &limit) {
			helpers.WriteJson(w, http.StatusTooManyRequests, map[string]any{
				"status":     false,
				"message":    limit.Message,
				"retryAfter": limit.RetryAfter,
			})
			return
		}

		log.Println(err)
		
		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "Something went wrong. Please try again.",
		})
		return
	}

	if err := app.Mail.SendVerificationCode(email, code, otp.CodeTTL, otp.MaxAttempts); err != nil {
		log.Println("SEND EMAIL:", err)

		app.OTP.Forget(email)

		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "could not send the email",
		})
		return
	}

	helpers.WriteJson(w, http.StatusOK, map[string]any{
		"status":   true,
		"message":  "Code sent",
		"attempts": otp.MaxAttempts,
		"cooldown": int(otp.Cooldown.Seconds()),
	})
}

func (app *App) VerifyEmailCode(w http.ResponseWriter, r *http.Request) {
	var req verifyCodeRequest

	r.Body = http.MaxBytesReader(w, r.Body, 1<<10)

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": "could not read request",
		})
		return
	}

	email := normalizeEmail(req.Email)
	code := strings.TrimSpace(req.Code)

	if !isSixDigits(code) {
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":  false,
			"message": fmt.Sprintf("Enter the %d-digit code", otp.CodeLength),
		})
		return
	}

	token, left, err := app.OTP.Verify(email, code)

	switch {
	case err == nil:
		helpers.WriteJson(w, http.StatusOK, map[string]any{
			"status": true,
			"token":  token,
		})

	case errors.Is(err, otp.ErrWrongCode):
		word := "tries"

		if left == 1 {
			word = "try"
		}

		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":       false,
			"message":      fmt.Sprintf("Wrong code. %d %s left.", left, word),
			"attemptsLeft": left,
		})

	case errors.Is(err, otp.ErrLocked):
		helpers.WriteJson(w, http.StatusTooManyRequests, map[string]any{
			"status":       false,
			"message":      "Too many wrong attempts. Request a new code.",
			"attemptsLeft": 0,
		})

	case errors.Is(err, otp.ErrExpired):
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":       false,
			"message":      "This code expired. Request a new one.",
			"attemptsLeft": 0,
		})

	case errors.Is(err, otp.ErrNoCode):
		helpers.WriteJson(w, http.StatusBadRequest, map[string]any{
			"status":       false,
			"message":      "No code found for this email. Request a new one.",
			"attemptsLeft": 0,
		})

	default:
		log.Println(err)

		helpers.WriteJson(w, http.StatusInternalServerError, map[string]any{
			"status":  false,
			"message": "Something went wrong. Please try again.",
		})
	}
}
