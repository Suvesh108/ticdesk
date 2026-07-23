package handlers

import (
	"html/template"
	"net/http"
	"ticDesk/internal/auth"
	"ticDesk/internal/models"
	"ticDesk/internal/repository"

	"github.com/go-chi/chi/v5"
)

type AdminHandler struct {
	userRepo *repository.UserRepository
}

func NewAdminHandler(userRepo *repository.UserRepository) *AdminHandler {
	return &AdminHandler{userRepo: userRepo}
}

type AdminUsersData struct {
	User  *models.User
	Users []models.User
}

func (h *AdminHandler) ShowUserManagement(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	users, err := h.userRepo.ListUsers(r.Context())
	if err != nil {
		http.Error(w, "Failed to load users: "+err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl, err := template.New("base.html").Funcs(template.FuncMap{
		"sub": func(a, b int) int { return a - b },
		"add": func(a, b int) int { return a + b },
		"mul": func(a, b int) int { return a * b },
	}).ParseFiles(
		"web/templates/layouts/base.html",
		"web/templates/pages/admin_users.html",
	)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Title": "User Management — ticDesk",
		"User":  user,
		"Users": users,
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tmpl.ExecuteTemplate(w, "base.html", data)
}

func (h *AdminHandler) UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	targetUserID := chi.URLParam(r, "id")
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form submission", http.StatusBadRequest)
		return
	}

	roleStr := r.FormValue("role")
	newRole := models.RoleCustomer
	if roleStr == "admin" {
		newRole = models.RoleAdmin
	} else if roleStr == "support" {
		newRole = models.RoleSupport
	}

	if err := h.userRepo.UpdateUserRole(r.Context(), targetUserID, newRole); err != nil {
		http.Error(w, "Failed to update user role", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

func (h *AdminHandler) ToggleUserStatus(w http.ResponseWriter, r *http.Request) {
	targetUserID := chi.URLParam(r, "id")
	if _, err := h.userRepo.ToggleUserStatus(r.Context(), targetUserID); err != nil {
		http.Error(w, "Failed to update user status", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}
