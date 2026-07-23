package handlers

import (
	"html/template"
	"log"
	"net/http"
	"strings"
	"ticDesk/internal/auth"
	"ticDesk/internal/models"
	"ticDesk/internal/repository"
)

type AuthHandler struct {
	userRepo *repository.UserRepository
	tmpl     *template.Template
}

func NewAuthHandler(userRepo *repository.UserRepository) *AuthHandler {
	tmpl, err := template.ParseFiles(
		"web/templates/layouts/base.html",
		"web/templates/pages/login.html",
		"web/templates/pages/register.html",
		"web/templates/partials/toast.html",
	)
	if err != nil {
		log.Printf("Warning: error parsing template files in AuthHandler: %v", err)
	}

	return &AuthHandler{
		userRepo: userRepo,
		tmpl:     tmpl,
	}
}

type AuthData struct {
	User  *models.User
	Email string
	Error string
}

func (h *AuthHandler) ShowLogin(w http.ResponseWriter, r *http.Request) {
	if auth.SessionManager.Exists(r.Context(), "user_id") {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	data := AuthData{}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.ExecuteTemplate(w, "login.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *AuthHandler) ProcessLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form submission", http.StatusBadRequest)
		return
	}

	email := strings.TrimSpace(strings.ToLower(r.FormValue("email")))
	password := r.FormValue("password")

	user, err := h.userRepo.GetUserByEmail(r.Context(), email)
	if err != nil || !auth.CheckPasswordHash(password, user.PasswordHash) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusUnauthorized)
		data := AuthData{Email: email, Error: "Invalid email or password"}
		_ = h.tmpl.ExecuteTemplate(w, "login.html", data)
		return
	}

	// Session renew & set
	if err := auth.SessionManager.RenewToken(r.Context()); err != nil {
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}

	auth.SessionManager.Put(r.Context(), "user_id", user.ID)
	auth.SessionManager.Put(r.Context(), "user_name", user.Name)
	auth.SessionManager.Put(r.Context(), "user_email", user.Email)
	auth.SessionManager.Put(r.Context(), "user_role", string(user.Role))

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (h *AuthHandler) ShowRegister(w http.ResponseWriter, r *http.Request) {
	if auth.SessionManager.Exists(r.Context(), "user_id") {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	data := AuthData{}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.ExecuteTemplate(w, "register.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *AuthHandler) ProcessRegister(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form submission", http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	email := strings.TrimSpace(strings.ToLower(r.FormValue("email")))
	password := r.FormValue("password")
	roleStr := r.FormValue("role")

	if name == "" || email == "" || len(password) < 6 {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		data := AuthData{Error: "Please provide valid registration details (password min 6 chars)"}
		_ = h.tmpl.ExecuteTemplate(w, "register.html", data)
		return
	}

	role := models.RoleCustomer
	if roleStr == "admin" {
		role = models.RoleAdmin
	} else if roleStr == "support" {
		role = models.RoleSupport
	}

	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		http.Error(w, "Error processing password", http.StatusInternalServerError)
		return
	}

	user, err := h.userRepo.CreateUser(r.Context(), name, email, hashedPassword, role)
	if err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusConflict)
		data := AuthData{Error: "An account with this email already exists"}
		_ = h.tmpl.ExecuteTemplate(w, "register.html", data)
		return
	}

	if err := auth.SessionManager.RenewToken(r.Context()); err != nil {
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}

	auth.SessionManager.Put(r.Context(), "user_id", user.ID)
	auth.SessionManager.Put(r.Context(), "user_name", user.Name)
	auth.SessionManager.Put(r.Context(), "user_email", user.Email)
	auth.SessionManager.Put(r.Context(), "user_role", string(user.Role))

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (h *AuthHandler) ProcessLogout(w http.ResponseWriter, r *http.Request) {
	_ = auth.SessionManager.Destroy(r.Context())
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
