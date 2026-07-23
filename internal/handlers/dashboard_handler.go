package handlers

import (
	"html/template"
	"log"
	"net/http"
	"ticDesk/internal/auth"
	"ticDesk/internal/models"
)

type DashboardHandler struct {
	tmpl *template.Template
}

func NewDashboardHandler() *DashboardHandler {
	tmpl, err := template.ParseFiles(
		"web/templates/layouts/base.html",
		"web/templates/pages/dashboard.html",
	)
	if err != nil {
		log.Printf("Warning: error parsing dashboard template: %v", err)
	}

	return &DashboardHandler{tmpl: tmpl}
}

type DashboardData struct {
	User *models.User
}

func (h *DashboardHandler) ShowDashboard(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	data := DashboardData{User: user}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.ExecuteTemplate(w, "dashboard.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
