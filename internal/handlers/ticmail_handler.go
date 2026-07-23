package handlers

import (
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"ticDesk/internal/auth"
	"ticDesk/internal/models"
	"ticDesk/internal/repository"
	"ticDesk/internal/services"
	"time"
)

type TicMailHandler struct {
	ticmailRepo  *repository.TicMailRepository
	userRepo     *repository.UserRepository
	ticketRepo   *repository.TicketRepository
	emailService *services.EmailService
	tmpl         *template.Template
}

func NewTicMailHandler(
	ticmailRepo *repository.TicMailRepository,
	userRepo *repository.UserRepository,
	ticketRepo *repository.TicketRepository,
	emailService *services.EmailService,
	tmpl *template.Template,
) *TicMailHandler {
	return &TicMailHandler{
		ticmailRepo:  ticmailRepo,
		userRepo:     userRepo,
		ticketRepo:   ticketRepo,
		emailService: emailService,
		tmpl:         tmpl,
	}
}

func (h *TicMailHandler) RenderTicMail(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	logs, err := h.ticmailRepo.ListLogs(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl, err := template.New("").Funcs(template.FuncMap{
		"sub": func(a, b int) int { return a - b },
		"add": func(a, b int) int { return a + b },
		"mul": func(a, b int) int { return a * b },
	}).ParseFiles(
		"web/templates/layouts/base.html",
		"web/templates/pages/ticmail.html",
	)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Title": "ticMail Alert — ticDesk",
		"User":  user,
		"Logs":  logs,
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tmpl.ExecuteTemplate(w, "base.html", data)
}

func (h *TicMailHandler) ClearTicMail(w http.ResponseWriter, r *http.Request) {
	_ = h.ticmailRepo.ClearLogs(r.Context())
	http.Redirect(w, r, "/ticmail", http.StatusSeeOther)
}

func (h *TicMailHandler) SimulateInboundEmail(w http.ResponseWriter, r *http.Request) {
	senderName := r.FormValue("sender_name")
	senderEmail := r.FormValue("sender_email")
	subject := r.FormValue("subject")
	description := r.FormValue("description")
	priorityStr := r.FormValue("priority")

	if senderEmail == "" || subject == "" || description == "" {
		http.Error(w, "Email, Subject, and Description are required", http.StatusBadRequest)
		return
	}

	priority := models.PriorityMedium
	if priorityStr == "high" {
		priority = models.PriorityHigh
	} else if priorityStr == "low" {
		priority = models.PriorityLow
	}

	// 1. Check if user exists or create temporary guest customer account
	existingUser, err := h.userRepo.GetUserByEmail(r.Context(), senderEmail)
	var userID string
	var tempPassword string

	if err != nil || existingUser == nil {
		// Generate random temporary password
		randSource := rand.New(rand.NewSource(time.Now().UnixNano()))
		tempPassword = fmt.Sprintf("tempPass_%d", randSource.Intn(899999)+100000)

		hashedPassword, _ := auth.HashPassword(tempPassword)
		if senderName == "" {
			senderName = "Guest Customer"
		}

		newUser, err := h.userRepo.CreateTemporaryUser(r.Context(), senderName, senderEmail, hashedPassword)
		if err != nil {
			http.Error(w, "Failed to create guest user: "+err.Error(), http.StatusInternalServerError)
			return
		}
		userID = newUser.ID
	} else {
		userID = existingUser.ID
	}

	// 2. Create ticket (Auto-assignment algorithm runs inside CreateTicket)
	ticket, err := h.ticketRepo.CreateTicket(r.Context(), subject, description, 1, priority, userID)
	if err != nil {
		http.Error(w, "Failed to create ticket: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 3. Send automated response email with temporary credentials & ticket number
	emailSubject := fmt.Sprintf("[ticMail Alert] Ticket #%d Created: %s", ticket.TicketNumber, ticket.Title)
	var emailBody string

	if tempPassword != "" {
		emailBody = fmt.Sprintf(`
			<div style="font-family: sans-serif; background-color: #faf9f8; padding: 16px; border-radius: 4px; border: 1px solid #edebe9;">
				<h3 style="color: #0078d4; margin-top: 0;">Welcome to ticDesk Support</h3>
				<p>Your support ticket <strong>#%d</strong> has been created and automatically assigned to support staff.</p>
				<div style="background-color: #deecf9; padding: 12px; border-radius: 4px; margin: 12px 0;">
					<p style="margin: 0; font-weight: bold; color: #005a9e;">Temporary Guest Credentials:</p>
					<p style="margin: 4px 0;">Email: <strong>%s</strong></p>
					<p style="margin: 4px 0;">Temporary Password: <strong>%s</strong></p>
					<p style="margin: 4px 0; font-size: 11px; color: #605e5c;">(Note: This temporary account will automatically delete upon ticket closure)</p>
				</div>
				<p><a href="http://localhost:8081/login" style="background-color: #0078d4; color: white; padding: 8px 16px; text-decoration: none; border-radius: 2px; font-weight: bold;">Log in to View Ticket</a></p>
			</div>
		`, ticket.TicketNumber, senderEmail, tempPassword)
	} else {
		emailBody = fmt.Sprintf(`
			<div style="font-family: sans-serif; background-color: #faf9f8; padding: 16px; border-radius: 4px; border: 1px solid #edebe9;">
				<h3 style="color: #0078d4; margin-top: 0;">Ticket Received</h3>
				<p>Your support ticket <strong>#%d</strong> has been created and automatically assigned to support staff.</p>
				<p><a href="http://localhost:8081/tickets/%s" style="background-color: #0078d4; color: white; padding: 8px 16px; text-decoration: none; border-radius: 2px; font-weight: bold;">View Ticket Status</a></p>
			</div>
		`, ticket.TicketNumber, ticket.ID)
	}

	h.emailService.Enqueue(services.EmailJob{
		To:      senderEmail,
		Subject: emailSubject,
		HTML:    emailBody,
	})

	http.Redirect(w, r, "/ticmail", http.StatusSeeOther)
}
