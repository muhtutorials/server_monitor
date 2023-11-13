package main

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"server_monitor/internal/handlers"
)

func routes() http.Handler {
	mux := chi.NewRouter()

	mux.Use(loadSession)
	mux.Use(recoverPanic)
	mux.Use(noSurf)
	mux.Use(checkForRememberMe)

	mux.Get("/", handlers.Repo.LoginScreen)
	mux.Post("/", handlers.Repo.Login)
	mux.Get("/user/logout", handlers.Repo.Logout)

	mux.Route("/pusher", func(mux chi.Router) {
		mux.Use(auth)
		mux.Post("/auth", handlers.Repo.PusherAuth)
	})

	mux.Route("/admin", func(mux chi.Router) {
		mux.Use(auth)

		// example of a private WS message
		mux.Get("/private-message", handlers.Repo.SendPrivateMessage)

		mux.Get("/dashboard", handlers.Repo.AdminDashboard)

		mux.Get("/events", handlers.Repo.Events)

		mux.Get("/settings", handlers.Repo.Settings)
		mux.Post("/settings", handlers.Repo.SaveSettings)

		mux.Get("/healthy", handlers.Repo.HealthyServices)
		mux.Get("/warning", handlers.Repo.WarningServices)
		mux.Get("/problem", handlers.Repo.ProblemServices)
		mux.Get("/pending", handlers.Repo.PendingServices)

		mux.Get("/users", handlers.Repo.UserList)
		mux.Get("/users/{id}", handlers.Repo.User)
		mux.Post("/users/{id}", handlers.Repo.CreateOrUpdateUser)
		mux.Post("/users/{id}/delete", handlers.Repo.DeleteUser)

		mux.Get("/schedule", handlers.Repo.TaskList)

		mux.Get("/hosts", handlers.Repo.HostList)
		mux.Get("/hosts/{id}", handlers.Repo.Host)
		mux.Post("/hosts/{id}", handlers.Repo.CreateOrUpdateHost)
		mux.Post("/hosts/toggle-host-is-active", handlers.Repo.ToggleHostIsActive)
		mux.Post("/hosts/toggle-service-is-active", handlers.Repo.ToggleServiceIsActive)
		mux.Get("/check-status/{id}/{oldStatus}", handlers.Repo.CheckStatus)
		mux.Post("/preferences", handlers.Repo.ToggleMonitoring)
	})

	// static files
	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	return mux
}
