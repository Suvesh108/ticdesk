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

func renderPartial(w http.ResponseWriter, partial string, data interface{}) {
	tmpl, err := template.ParseFiles("web/templates/partials/" + partial)
	if err != nil {
		http.Error(w, "Partial Error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Render Partial Error: "+err.Error(), http.StatusInternalServerError)
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
	Agents []models.User
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

	if user.Role == models.RoleCustomer && ticket.CreatedByID != user.ID {
		http.Error(w, "Forbidden: you cannot access this ticket", http.StatusForbidden)
		return
	}

	agents, _ := h.ticketRepo.GetSupportAgents(r.Context())

	data := TicketDetailData{
		User:   user,
		Ticket: ticket,
		Agents: agents,
	}
	renderTicketPage(w, "ticket_detail.html", data)
}

// HTMX Partial Update: Status
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

	if statusStr == "" {
		http.Error(w, "Status parameter required", http.StatusBadRequest)
		return
	}

	newStatus := models.TicketStatus(statusStr)

	updatedTicket, err := h.ticketRepo.UpdateTicketStatus(r.Context(), ticketID, newStatus, user.ID)
	if err != nil {
		http.Error(w, "Failed to update status: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := TicketDetailData{User: user, Ticket: updatedTicket}
	renderPartial(w, "ticket_status_badge.html", data)
}

// HTMX Partial Update: Priority
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

	if priorityStr == "" {
		http.Error(w, "Priority parameter required", http.StatusBadRequest)
		return
	}

	newPriority := models.TicketPriority(priorityStr)

	updatedTicket, err := h.ticketRepo.UpdateTicketPriority(r.Context(), ticketID, newPriority)
	if err != nil {
		http.Error(w, "Failed to update priority: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := TicketDetailData{User: user, Ticket: updatedTicket}
	renderPartial(w, "ticket_priority_badge.html", data)
}

// HTMX Partial Update: Assignee
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

	updatedTicket, err := h.ticketRepo.UpdateTicketAssignee(r.Context(), ticketID, assigneeID)
	if err != nil {
		http.Error(w, "Failed to update assignee: "+err.Error(), http.StatusInternalServerError)
		return
	}

	agents, _ := h.ticketRepo.GetSupportAgents(r.Context())
	data := TicketDetailData{User: user, Ticket: updatedTicket, Agents: agents}
	renderPartial(w, "ticket_assignee.html", data)
}
