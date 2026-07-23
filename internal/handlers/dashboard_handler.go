package handlers

import (
	"encoding/json"
	"html/template"
	"net/http"
	"ticDesk/internal/auth"
	"ticDesk/internal/models"
	"ticDesk/internal/repository"
)

type DashboardHandler struct {
	ticketRepo *repository.TicketRepository
}

func NewDashboardHandler(ticketRepo *repository.TicketRepository) *DashboardHandler {
	return &DashboardHandler{ticketRepo: ticketRepo}
}

type DashboardData struct {
	User    *models.User
	Stats   *models.DashboardStats
	Tickets []models.Ticket
}

func (h *DashboardHandler) ShowDashboard(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	stats, _ := h.ticketRepo.GetDashboardStats(r.Context(), user)
	recentTickets, _ := h.ticketRepo.ListTickets(r.Context(), user)
	if len(recentTickets) > 5 {
		recentTickets = recentTickets[:5]
	}

	tmpl, err := template.New("").Funcs(template.FuncMap{
		"sub": func(a, b int) int { return a - b },
		"add": func(a, b int) int { return a + b },
		"mul": func(a, b int) int { return a * b },
	}).ParseFiles(
		"web/templates/layouts/base.html",
		"web/templates/pages/dashboard.html",
	)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Title":   "Dashboard — ticDesk",
		"User":    user,
		"Stats":   stats,
		"Tickets": recentTickets,
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tmpl.ExecuteTemplate(w, "base.html", data)
}

func (h *DashboardHandler) GetStatsJSON(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	stats, err := h.ticketRepo.GetDashboardStats(r.Context(), user)
	if err != nil {
		http.Error(w, "Failed to load stats: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(stats)
}
