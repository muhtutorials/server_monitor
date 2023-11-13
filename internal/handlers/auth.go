package handlers

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"server_monitor/internal/helpers"
	"server_monitor/internal/models"
	"strings"
	"time"
)

func (repo *DBRepo) LoginScreen(w http.ResponseWriter, r *http.Request) {
	if repo.App.Session.Exists(r.Context(), "userID") {
		http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
		return
	}

	err := helpers.RenderPage(w, r, "login", nil, nil)
	if err != nil {
		printTemplateError(w, err)
	}
}

func (repo *DBRepo) Login(w http.ResponseWriter, r *http.Request) {
	_ = repo.App.Session.RenewToken(r.Context())

	err := r.ParseForm()
	if err != nil {
		app.ErrorLog.Println(err)
		ClientError(w, r, http.StatusBadRequest)
		return
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	userID, hashedPassword, err := repo.DB.Authenticate(email, password)
	if err == models.ErrInvalidCredentials {
		app.Session.Put(r.Context(), "error", "Invalid login")
		err = helpers.RenderPage(w, r, "login", nil, nil)
		if err != nil {
			printTemplateError(w, err)
		}
		return
	} else if err == models.ErrInactiveAccount {
		app.Session.Put(r.Context(), "error", "Inactive login")
		err = helpers.RenderPage(w, r, "login", nil, nil)
		if err != nil {
			printTemplateError(w, err)
		}
		return
	} else if err != nil {
		app.ErrorLog.Println(err)
		ClientError(w, r, http.StatusBadRequest)
		return
	}

	if r.Form.Get("remember") == "remember" {
		randomString := helpers.RandomString(12)

		hasher := sha256.New()
		_, err = hasher.Write([]byte(randomString))
		if err != nil {
			app.ErrorLog.Println(err)
		}

		token := base64.URLEncoding.EncodeToString(hasher.Sum(nil))

		err = repo.DB.InsertRememberMeToken(userID, token)
		if err != nil {
			app.ErrorLog.Println(err)
		}

		expires := time.Now().Add(365 * 24 * 60 * 60 * time.Second)
		cookie := http.Cookie{
			Name:     fmt.Sprintf("%s_remember_me", app.Preferences["identifier"]),
			Value:    fmt.Sprintf("%d|%s", userID, token),
			Path:     "/",
			Expires:  expires,
			HttpOnly: true,
			Domain:   app.Domain,
			MaxAge:   365 * 24 * 60 * 60,
			Secure:   app.InProduction,
			SameSite: http.SameSiteStrictMode,
		}
		http.SetCookie(w, &cookie)
	}

	user, err := repo.DB.GetUserByID(userID)
	if err != nil {
		app.ErrorLog.Println(err)
		ClientError(w, r, http.StatusBadRequest)
		return
	}

	app.Session.Put(r.Context(), "userID", userID)
	app.Session.Put(r.Context(), "hashedPassword", hashedPassword)
	app.Session.Put(r.Context(), "flash", "You've been logged in successfully!")
	app.Session.Put(r.Context(), "user", user)

	if r.Form.Get("target") != "" {
		http.Redirect(w, r, r.Form.Get("target"), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
}

func (repo *DBRepo) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(fmt.Sprintf("%s_remember_me", app.Preferences["identifier"]))
	if err == nil {
		key := cookie.Value
		if len(key) > 0 {
			split := strings.Split(key, "|")
			token := split[1]
			err = repo.DB.DeleteRememberMeToken(token)
			if err != nil {
				app.ErrorLog.Println(err)
			}
		}
	}

	deleteCookie := http.Cookie{
		Name:     fmt.Sprintf("%s_remember_me", app.Preferences["identifier"]),
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Domain:   app.Domain,
		MaxAge:   0,
	}
	http.SetCookie(w, &deleteCookie)

	_ = app.Session.RenewToken(r.Context())
	_ = app.Session.Destroy(r.Context())
	_ = app.Session.RenewToken(r.Context())

	repo.App.Session.Put(r.Context(), "flash", "You've been logged out successfully!")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
