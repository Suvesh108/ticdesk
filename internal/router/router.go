package router

import (
	"net/http"
	"ticDesk/internal/auth"
	"ticDesk/internal/handlers"
	"ticDesk/internal/models"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func New(
	sessionManager *scs.SessionManager,
	authHandler *handlers.AuthHandler,
	dashboardHandler *handlers.DashboardHandler,
	ticketHandler *handlers.TicketHandler,
	commentHandler *handlers.CommentHandler,
	attachmentHandler *handlers.AttachmentHandler,
	adminHandler *handlers.AdminHandler,
	calendarHandler *handlers.CalendarHandler,
	noteHandler *handlers.NoteHandler,
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

	// Protected Routes (Requires Authentication)
	r.Group(func(r chi.Router) {
		r.Use(auth.RequireAuth)

		r.Get("/dashboard", dashboardHandler.ShowDashboard)
		r.Get("/dashboard/stats.json", dashboardHandler.GetStatsJSON)

		// Tickets Routes (Role-Filtered inside repository queries)
		r.Get("/tickets", ticketHandler.ShowTicketList)
		r.Get("/tickets/new", ticketHandler.ShowNewTicket)
		r.Post("/tickets", ticketHandler.ProcessCreateTicket)
		r.Get("/tickets/{id}", ticketHandler.ShowTicketDetail)

		// Staff & Admin Only Routes (Calendar & Notes)
		r.Group(func(r chi.Router) {
			r.Use(auth.RequireRole(models.RoleAdmin, models.RoleSupport))

			r.Get("/calendar", calendarHandler.RenderCalendar)
			r.Post("/calendar/events", calendarHandler.CreateEvent)
			r.Post("/calendar/events/{id}/delete", calendarHandler.DeleteEvent)

			r.Get("/notes", noteHandler.RenderNotes)
			r.Post("/notes", noteHandler.CreateNote)
			r.Post("/notes/{id}/pin", noteHandler.TogglePin)
			r.Post("/notes/{id}/delete", noteHandler.DeleteNote)

			// HTMX Partial Mutation Routes
			r.Patch("/tickets/{id}/status", ticketHandler.UpdateStatus)
			r.Patch("/tickets/{id}/priority", ticketHandler.UpdatePriority)
			r.Patch("/tickets/{id}/assign", ticketHandler.UpdateAssignee)
		})

		// Comments & Attachments Routes
		r.Get("/tickets/{id}/comments", commentHandler.GetComments)
		r.Post("/tickets/{id}/comments", commentHandler.CreateComment)
		r.Get("/attachments/{id}", attachmentHandler.DownloadAttachment)

		// Admin-Only Routes (Requires Role: Admin)
		r.Group(func(r chi.Router) {
			r.Use(auth.RequireRole(models.RoleAdmin))

			r.Get("/admin/users", adminHandler.ShowUserManagement)
			r.Post("/admin/users/{id}/role", adminHandler.UpdateUserRole)
			r.Post("/admin/users/{id}/deactivate", adminHandler.ToggleUserStatus)
		})
	})

	return r
}
