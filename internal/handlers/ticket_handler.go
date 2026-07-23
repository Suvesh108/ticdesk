package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"ticDesk/internal/auth"
	"ticDesk/internal/models"
	"ticDesk/internal/repository"
	"ticDesk/internal/services"
)

type TicketHandler struct {
	repo         *repository.TicketRepository
	emailService *services.EmailService
}

func NewTicketHandler(repo *repository.TicketRepository, emailService *services.EmailService) *TicketHandler {
	return &TicketHandler{
		repo:         repo,
		emailService: emailService,
	}
}

type TicketListData struct {
	User       *models.User
	Tickets    []models.Ticket
	Categories []models.Category
	TotalCount int
	Page       int
	Limit      int
}

type TicketDetailData struct {
	User   *models.User
	Ticket *models.Ticket
	Agents []models.User
}

func (h *TicketHandler) ShowTicketList(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	search := r.URL.Query().Get("search")
	status := r.URL.Query().Get("status")
	priority := r.URL.Query().Get("priority")
	categoryIDStr := r.URL.Query().Get("category_id")
	pageStr := r.URL.Query().Get("page")

	categoryID, _ := strconv.Atoi(categoryIDStr)
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	limit := 10

	opts := repository.TicketFilterOptions{
		Search:     search,
		CategoryID: categoryID,
		Status:     models.TicketStatus(status),
		Priority:   models.TicketPriority(priority),
		Page:       page,
		Limit:      limit,
	}

	tickets, totalCount, err := h.repo.ListTicketsFiltered(r.Context(), user, opts)
	if err != nil {
		http.Error(w, "Failed to load tickets: "+err.Error(), http.StatusInternalServerError)
		return
	}

	categories, _ := h.repo.GetCategories(r.Context())

	data := TicketListData{
		User:       user,
		Tickets:    tickets,
		Categories: categories,
		TotalCount: totalCount,
		Page:       page,
		Limit:      limit,
	}

	// Check if this is an HTMX request
	if r.Header.Get("HX-Request") == "true" {
		tmpl, err := template.ParseFiles("web/templates/partials/ticket_table.html")
		if err != nil {
			http.Error(w, "Template Error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = tmpl.Execute(w, data)
		return
	}

	tmpl, err := template.ParseFiles(
		"web/templates/layouts/base.html",
		"web/templates/pages/ticket_list.html",
		"web/templates/partials/ticket_table.html",
	)
	if err != nil {
		http.Error(w, "Template Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tmpl.ExecuteTemplate(w, "base.html", data)
}

func (h *TicketHandler) ShowNewTicket(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	categories, err := h.repo.GetCategories(r.Context())
	if err != nil {
		http.Error(w, "Failed to load categories: "+err.Error(), http.StatusInternalServerError)
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

	data := map[string]interface{}{
		"User":       user,
		"Categories": categories,
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tmpl.ExecuteTemplate(w, "base.html", data)
}

func (h *TicketHandler) ProcessCreateTicket(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	_ = r.ParseForm()
	title := r.FormValue("title")
	description := r.FormValue("description")
	categoryIDStr := r.FormValue("category_id")
	priorityStr := r.FormValue("priority")

	if title == "" || description == "" {
		http.Error(w, "Title and description are required", http.StatusBadRequest)
		return
	}

	categoryID, _ := strconv.Atoi(categoryIDStr)
	priority := models.TicketPriority(priorityStr)
	if priority == "" {
		priority = models.PriorityMedium
	}

	ticket, err := h.repo.CreateTicket(r.Context(), title, description, categoryID, priority, user.ID)
	if err != nil {
		http.Error(w, "Failed to create ticket: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Trigger Email Notification Worker
	if h.emailService != nil {
		h.emailService.NotifyTicketCreated(ticket, user.Email)
	}

	http.Redirect(w, r, fmt.Sprintf("/tickets/%s", ticket.ID), http.StatusSeeOther)
}

func (h *TicketHandler) ShowTicketDetail(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	ticketID := r.PathValue("id")
	if ticketID == "" {
		http.NotFound(w, r)
		return
	}

	ticket, err := h.repo.GetTicketByID(r.Context(), ticketID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	agents, _ := h.repo.GetSupportAgents(r.Context())

	tmpl, err := template.ParseFiles(
		"web/templates/layouts/base.html",
		"web/templates/pages/ticket_detail.html",
	)
	if err != nil {
		http.Error(w, "Template Error: "+err.Error(), http.StatusInternalServerError)
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
	if user == nil || (user.Role != models.RoleAdmin && user.Role != models.RoleSupport) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	_ = r.ParseForm()
	ticketID := r.PathValue("id")
	newStatusStr := r.FormValue("status")
	if newStatusStr == "" {
		newStatusStr = r.URL.Query().Get("status")
	}

	newStatus := models.TicketStatus(newStatusStr)
	ticket, err := h.repo.UpdateTicketStatus(r.Context(), ticketID, newStatus, user.ID)
	if err != nil {
		http.Error(w, "Failed to update status: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Trigger Email Alert
	if h.emailService != nil {
		h.emailService.NotifyStatusChanged(ticket, ticket.Status, newStatus, user.Email)
	}

	tmpl, err := template.ParseFiles("web/templates/partials/ticket_status_badge.html")
	if err != nil {
		http.Error(w, "Template Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"User":   user,
		"Ticket": ticket,
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tmpl.Execute(w, data)
}

func (h *TicketHandler) UpdatePriority(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil || (user.Role != models.RoleAdmin && user.Role != models.RoleSupport) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	_ = r.ParseForm()
	ticketID := r.PathValue("id")
	newPriorityStr := r.FormValue("priority")
	newPriority := models.TicketPriority(newPriorityStr)

	ticket, err := h.repo.UpdateTicketPriority(r.Context(), ticketID, newPriority)
	if err != nil {
		http.Error(w, "Failed to update priority: "+err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("web/templates/partials/ticket_priority_badge.html")
	if err != nil {
		http.Error(w, "Template Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"User":   user,
		"Ticket": ticket,
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tmpl.Execute(w, data)
}

func (h *TicketHandler) UpdateAssignee(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil || (user.Role != models.RoleAdmin && user.Role != models.RoleSupport) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	_ = r.ParseForm()
	ticketID := r.PathValue("id")
	assigneeIDStr := r.FormValue("assigned_to")

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

	tmpl, err := template.ParseFiles("web/templates/partials/ticket_assignee.html")
	if err != nil {
		http.Error(w, "Template Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"User":   user,
		"Ticket": ticket,
		"Agents": agents,
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tmpl.Execute(w, data)
}
