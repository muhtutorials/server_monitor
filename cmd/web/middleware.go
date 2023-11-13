package main

import (
	"fmt"
	"github.com/justinas/nosurf"
	"net/http"
	"server_monitor/internal/helpers"
	"strconv"
	"strings"
	"time"
)

func loadSession(next http.Handler) http.Handler {
	return session.LoadAndSave(next)
}

func recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				helpers.ServerError(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func noSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)

	csrfHandler.ExemptPaths("/pusher/auth", "/pusher/hook")

	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   app.InProduction,
		SameSite: http.SameSiteStrictMode,
		Domain:   app.Domain,
	})

	return csrfHandler
}

// checkForRemember checks to see if we should log the user in automatically
func checkForRememberMe(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !helpers.IsAuthenticated(r) {
			cookie, err := r.Cookie(fmt.Sprintf("%s_remember_me", preferences["identifier"]))
			if err != nil {
				// no cookie
				next.ServeHTTP(w, r)
			} else {
				key := cookie.Value
				if len(key) > 0 {
					// key length > 0, so it might be a valid token
					split := strings.Split(key, "|")
					uid, hash := split[0], split[1]
					id, _ := strconv.Atoi(uid)
					isValid := repo.DB.CheckForRememberMeToken(id, hash)
					if isValid {
						// valid remember me token, so log the user in
						_ = session.RenewToken(r.Context())
						user, _ := repo.DB.GetUserByID(id)
						session.Put(r.Context(), "userID", id)
						session.Put(r.Context(), "userName", user.FirstName)
						session.Put(r.Context(), "userFirstName", user.FirstName)
						session.Put(r.Context(), "userLastName", user.LastName)
						session.Put(r.Context(), "hashedPassword", string(user.Password))
						session.Put(r.Context(), "user", user)
						next.ServeHTTP(w, r)
					} else {
						// invalid token, so delete the cookie
						deleteRememberMeCookie(w, r)
						session.Put(r.Context(), "error", "You've been logged out from another device!")
						next.ServeHTTP(w, r)
					}
				} else {
					// key length is zero, so it's a leftover cookie (user has not closed browser)
					next.ServeHTTP(w, r)
				}
			}
		} else {
			// they are logged in, but make sure that the remember token has not been revoked
			cookie, err := r.Cookie(fmt.Sprintf("%s_remember_me", preferences["identifier"]))
			if err != nil {
				// no cookie
				next.ServeHTTP(w, r)
			} else {
				key := cookie.Value
				if len(key) > 0 {
					// key length > 0, so it might be a valid token
					split := strings.Split(key, "|")
					uid, hash := split[0], split[1]
					id, _ := strconv.Atoi(uid)
					isValid := repo.DB.CheckForRememberMeToken(id, hash)
					if !isValid {
						deleteRememberMeCookie(w, r)
						session.Put(r.Context(), "error", "You've been logged out from another device!")
						next.ServeHTTP(w, r)
					} else {
						next.ServeHTTP(w, r)
					}
				} else {
					next.ServeHTTP(w, r)
				}
			}
		}
	})
}

func deleteRememberMeCookie(w http.ResponseWriter, r *http.Request) {
	_ = session.RenewToken(r.Context())

	// deletes the cookie
	newCookie := http.Cookie{
		Name:     fmt.Sprintf("%s_remember_me", preferences["identifier"]),
		Path:     "/",
		Expires:  time.Now().Add(-100 * time.Hour),
		HttpOnly: true,
		Domain:   app.Domain,
		MaxAge:   -1,
		Secure:   app.InProduction,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, &newCookie)

	// log them out
	session.Remove(r.Context(), "userID")
	_ = session.Destroy(r.Context())
	_ = session.RenewToken(r.Context())
}

func auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !helpers.IsAuthenticated(r) {
			url := r.URL.Path
			http.Redirect(w, r, fmt.Sprintf("/?target=%s", url), http.StatusSeeOther)
			return
		}

		w.Header().Add("Cache-Control", "no-store")

		next.ServeHTTP(w, r)
	})
}
