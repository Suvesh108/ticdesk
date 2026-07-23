package handlers

import (
	"html/template"
	"net/http"
	"strconv"

	"ticDesk/internal/auth"
	"ticDesk/internal/models"
	"ticDesk/internal/repository"
	"ticDesk/internal/services"

	"github.com/go-chi/chi/v5"
)

type TicketHandler struct {
	repo         *repository.TicketRepository
	emailService *services.EmailService
}

func NewTicketHandler(repo *repository.TicketRepository, emailService *services.EmailService) *TicketHandler {
	return &TicketHandler{repo: repo, emailService: emailService}
}

type TicketListData struct {
	User    *models.User
	Tickets []models.Ticket
}

type TicketNewData struct {
	User       *models.User
	Categories []models.Category
}

type TicketDetailData struct {
	User   *models.User
	Ticket *models.Ticket
	Agents []models.User
}

func (h *TicketHandler) ShowTicketList(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	tickets, err := h.repo.ListTickets(r.Context(), user)
	if err != nil {
		http.Error(w, "Failed to load tickets: "+err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles(
		"web/templates/layouts/base.html",
		"web/templates/pages/ticket_list.html",
	)
	if err != nil {
		http.Error(w, "Template Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := TicketListData{User: user, Tickets: tickets}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tmpl.ExecuteTemplate(w, "base.html", data)
}

func (h *TicketHandler) ShowNewTicket(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	categories, err := h.repo.GetCategories(r.Context())
	if err != nil {
		http.Error(w, "Failed to load categories", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles(
		"web/templates/layouts/base.html",
		"web/templates/pages/ticket_new.html",
	)
	if err != nil {
		http.Error(w, "Template Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := TicketNewData{User: user, Categories: categories}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tmpl.ExecuteTemplate(w, "base.html", data)
}

func (h *TicketHandler) ProcessCreateTicket(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form submission", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	description := r.FormValue("description")
	categoryIDStr := r.FormValue("category_id")
	priorityStr := r.FormValue("priority")

	categoryID, err := strconv.Atoi(categoryIDStr)
	if err != nil {
		http.Error(w, "Invalid category", http.StatusBadRequest)
		return
	}

	priority := models.TicketPriority(priorityStr)
	if priority != models.PriorityLow && priority != models.PriorityMedium && priority != models.PriorityHigh {
		priority = models.PriorityMedium
	}

	ticket, err := h.repo.CreateTicket(r.Context(), title, description, categoryID, priority, user.ID)
	if err != nil {
		http.Error(w, "Failed to create ticket: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Trigger Email Notification (non-blocking)
	if h.emailService != nil {
		h.emailService.NotifyTicketCreated(ticket, user.Email)
	}

	http.Redirect(w, r, "/tickets/"+ticket.ID, http.StatusSeeOther)
}

func (h *TicketHandler) ShowTicketDetail(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	ticketID := chi.URLParam(r, "id")

	ticket, err := h.repo.GetTicketByID(r.Context(), ticketID)
	if err != nil {
		http.Error(w, "Ticket not found", http.StatusNotFound)
		return
	}

	if user.Role == models.RoleCustomer && ticket.CreatedByID != user.ID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	agents, _ := h.repo.GetSupportAgents(r.Context())

	tmpl, err := template.ParseFiles(
		"web/templates/layouts/base.html",
		"web/templates/pages/ticket_detail.html",
	)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := TicketDetailData{
		User:   user,
		Ticket: ticket,
		Agents: agents,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tmpl.ExecuteTemplate(w, "base.html", data)
}

func (h *TicketHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	ticketID := chi.URLParam(r, "id")

	if user.Role != models.RoleAdmin && user.Role != models.RoleSupport {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	_ = r.ParseForm()
	statusStr := r.FormValue("status")
	if statusStr == "" {
		statusStr = r.URL.Query().Get("status")
	}

	newStatus := models.TicketStatus(statusStr)
	ticket, err := h.repo.UpdateTicketStatus(r.Context(), ticketID, newStatus, user.ID)
	if err != nil {
		http.Error(w, "Failed to update status: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Trigger Email Notification (non-blocking)
	if h.emailService != nil {
		h.emailService.NotifyStatusChanged(ticket, ticket.Status, newStatus, user.Email)
	}

	renderPartial(w, "ticket_status_badge.html", map[string]interface{}{"Ticket": ticket, "User": user})
}

func (h *TicketHandler) UpdatePriority(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	ticketID := chi.URLParam(r, "id")

	if user.Role != models.RoleAdmin && user.Role != models.RoleSupport {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	_ = r.ParseForm()
	priorityStr := r.FormValue("priority")
	if priorityStr == "" {
		priorityStr = r.URL.Query().Get("priority")
	}

	newPriority := models.TicketPriority(priorityStr)
	ticket, err := h.repo.UpdateTicketPriority(r.Context(), ticketID, newPriority)
	if err != nil {
		http.Error(w, "Failed to update priority: "+err.Error(), http.StatusInternalServerError)
		return
	}

	renderPartial(w, "ticket_priority_badge.html", map[string]interface{}{"Ticket": ticket, "User": user})
}

func (h *TicketHandler) UpdateAssignee(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	ticketID := chi.URLParam(r, "id")

	if user.Role != models.RoleAdmin && user.Role != models.RoleSupport {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	_ = r.ParseForm()
	assigneeIDStr := r.FormValue("assigned_to")
	if assigneeIDStr == "" {
		assigneeIDStr = r.URL.Query().Get("assigned_to")
	}

	var assigneeID *string
	if assigneeIDStr != "" {
		assigneeID = &assigneeIDStr
	}

	ticket, err := h.repo.UpdateTicketAssignee(r.Context(), ticketID, assigneeID)
	if err != nil {
		http.Error(w, "Failed to update assignee: "+err.Error(), http.StatusInternalServerError)
		return
	}

	agents, _ := h.repo.GetSupportAgents(r.Context())
	renderPartial(w, "ticket_assignee.html", map[string]interface{}{"Ticket": ticket, "User": user, "Agents": agents})
}

func renderPartial(w http.ResponseWriter, templateName string, data interface{}) {
	tmpl, err := template.ParseFiles("web/templates/partials/" + templateName)
	if err != nil {
		http.Error(w, "Partial render error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tmpl.Execute(w, data)
}
