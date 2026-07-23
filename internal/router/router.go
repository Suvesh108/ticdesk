package router

import (
	"net/http"
	"ticDesk/internal/auth"
	"ticDesk/internal/handlers"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func New(
	sessionManager *scs.SessionManager,
	authHandler *handlers.AuthHandler,
	dashboardHandler *handlers.DashboardHandler,
) http.Handler {
	r := chi.NewRouter()

	// Global Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(sessionManager.LoadAndSave)

	// Root redirect
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		if sessionManager.Exists(r.Context(), "user_id") {
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		} else {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
		}
	})

	// Public Auth Routes
	r.Get("/login", authHandler.ShowLogin)
	r.Post("/login", authHandler.ProcessLogin)
	r.Get("/register", authHandler.ShowRegister)
	r.Post("/register", authHandler.ProcessRegister)
	r.Post("/logout", authHandler.ProcessLogout)

	// Protected Routes
	r.Group(func(r chi.Router) {
		r.Use(auth.RequireAuth)
		r.Get("/dashboard", dashboardHandler.ShowDashboard)
	})

	return r
}
