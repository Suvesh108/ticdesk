package handlers

import (
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"ticDesk/internal/auth"
	"ticDesk/internal/models"
	"ticDesk/internal/repository"

	"github.com/go-chi/chi/v5"
)

type TicketHandler struct {
	ticketRepo *repository.TicketRepository
}

func NewTicketHandler(ticketRepo *repository.TicketRepository) *TicketHandler {
	return &TicketHandler{
		ticketRepo: ticketRepo,
	}
}

func renderTicketPage(w http.ResponseWriter, page string, data interface{}) {
	tmpl, err := template.ParseFiles(
		"web/templates/layouts/base.html",
		"web/templates/pages/"+page,
		"web/templates/partials/toast.html",
	)
	if err != nil {
		http.Error(w, "Template Error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "base.html", data); err != nil {
		http.Error(w, "Render Error: "+err.Error(), http.StatusInternalServerError)
	}
}

type TicketListData struct {
	User    *models.User
	Tickets []models.Ticket
}

type TicketNewData struct {
	User       *models.User
	Categories []models.Category
	Error      string
}

type TicketDetailData struct {
	User   *models.User
	Ticket *models.Ticket
}

func (h *TicketHandler) ShowTicketList(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	tickets, err := h.ticketRepo.ListTickets(r.Context(), user)
	if err != nil {
		http.Error(w, "Failed to load tickets: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := TicketListData{
		User:    user,
		Tickets: tickets,
	}
	renderTicketPage(w, "ticket_list.html", data)
}

func (h *TicketHandler) ShowNewTicket(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	categories, err := h.ticketRepo.GetCategories(r.Context())
	if err != nil {
		http.Error(w, "Failed to load categories", http.StatusInternalServerError)
		return
	}

	data := TicketNewData{
		User:       user,
		Categories: categories,
	}
	renderTicketPage(w, "ticket_new.html", data)
}

func (h *TicketHandler) ProcessCreateTicket(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form submission", http.StatusBadRequest)
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))
	description := strings.TrimSpace(r.FormValue("description"))
	categoryIDStr := r.FormValue("category_id")
	priorityStr := r.FormValue("priority")

	categoryID, _ := strconv.Atoi(categoryIDStr)
	priority := models.PriorityMedium
	if priorityStr == "low" {
		priority = models.PriorityLow
	} else if priorityStr == "high" {
		priority = models.PriorityHigh
	}

	if title == "" || description == "" || categoryID == 0 {
		categories, _ := h.ticketRepo.GetCategories(r.Context())
		data := TicketNewData{
			User:       user,
			Categories: categories,
			Error:      "Please fill in all required fields",
		}
		w.WriteHeader(http.StatusBadRequest)
		renderTicketPage(w, "ticket_new.html", data)
		return
	}

	ticket, err := h.ticketRepo.CreateTicket(r.Context(), title, description, categoryID, priority, user.ID)
	if err != nil {
		http.Error(w, "Failed to create ticket: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/tickets/"+ticket.ID, http.StatusSeeOther)
}

func (h *TicketHandler) ShowTicketDetail(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	ticketID := chi.URLParam(r, "id")

	ticket, err := h.ticketRepo.GetTicketByID(r.Context(), ticketID)
	if err != nil {
		http.Error(w, "Ticket not found", http.StatusNotFound)
		return
	}

	// RBAC check: Customer can only view own tickets
	if user.Role == models.RoleCustomer && ticket.CreatedByID != user.ID {
		http.Error(w, "Forbidden: you cannot access this ticket", http.StatusForbidden)
		return
	}

	data := TicketDetailData{
		User:   user,
		Ticket: ticket,
	}
	renderTicketPage(w, "ticket_detail.html", data)
}
