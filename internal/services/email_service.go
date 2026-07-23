package services

import (
	"context"
	"fmt"
	"log"
	"net/smtp"
	"strings"
	"ticDesk/internal/models"
)

type EmailJob struct {
	To      string
	Subject string
	HTML    string
}

type EmailService struct {
	smtpHost string
	smtpPort string
	fromAddr string
	jobChan  chan EmailJob
}

func NewEmailService(host, port, from string) *EmailService {
	return &EmailService{
		smtpHost: host,
		smtpPort: port,
		fromAddr: from,
		jobChan:  make(chan EmailJob, 100),
	}
}

func (s *EmailService) StartWorker(ctx context.Context) {
	log.Printf("Email worker started listening on %s:%s", s.smtpHost, s.smtpPort)
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("Email worker stopping...")
				return
			case job, ok := <-s.jobChan:
				if !ok {
					return
				}
				if err := s.sendSMTP(job); err != nil {
					log.Printf("Failed to send email to %s: %v", job.To, err)
				} else {
					log.Printf("Successfully sent notification email to %s [%s]", job.To, job.Subject)
				}
			}
		}
	}()
}

func (s *EmailService) Enqueue(job EmailJob) {
	select {
	case s.jobChan <- job:
	default:
		log.Printf("Warning: Email queue full, dropping notification for %s", job.To)
	}
}

func (s *EmailService) sendSMTP(job EmailJob) error {
	addr := fmt.Sprintf("%s:%s", s.smtpHost, s.smtpPort)

	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("ticDesk Support <%s>", s.fromAddr)
	headers["To"] = job.To
	headers["Subject"] = job.Subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	var msg strings.Builder
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n")
	msg.WriteString(job.HTML)

	// Send via unauthenticated SMTP (MailHog)
	return smtp.SendMail(addr, nil, s.fromAddr, []string{job.To}, []byte(msg.String()))
}

// Notification Helpers

func (s *EmailService) NotifyTicketCreated(ticket *models.Ticket, recipientEmail string) {
	subject := fmt.Sprintf("[ticDesk] Ticket #%d Created: %s", ticket.TicketNumber, ticket.Title)
	body := fmt.Sprintf(`
		<div style="font-family: sans-serif; background-color: #0f172a; color: #f8fafc; padding: 24px; borderRadius: 12px;">
			<h2 style="color: #6366f1;">Ticket #%d Created</h2>
			<p>A new support ticket has been submitted on <strong>ticDesk</strong>.</p>
			<div style="background-color: #1e293b; padding: 16px; border-radius: 8px; margin: 16px 0;">
				<p><strong>Title:</strong> %s</p>
				<p><strong>Priority:</strong> %s</p>
				<p><strong>Status:</strong> %s</p>
			</div>
			<p><a href="http://localhost:8081/tickets/%s" style="background-color: #6366f1; color: #ffffff; padding: 10px 20px; text-decoration: none; border-radius: 8px; font-weight: bold;">View Ticket</a></p>
		</div>
	`, ticket.TicketNumber, ticket.Title, ticket.Priority, ticket.Status, ticket.ID)

	s.Enqueue(EmailJob{To: recipientEmail, Subject: subject, HTML: body})
}

func (s *EmailService) NotifyStatusChanged(ticket *models.Ticket, oldStatus, newStatus models.TicketStatus, recipientEmail string) {
	subject := fmt.Sprintf("[ticDesk] Ticket #%d Status Updated: %s -> %s", ticket.TicketNumber, oldStatus, newStatus)
	body := fmt.Sprintf(`
		<div style="font-family: sans-serif; background-color: #0f172a; color: #f8fafc; padding: 24px; borderRadius: 12px;">
			<h2 style="color: #6366f1;">Ticket Status Updated</h2>
			<p>Ticket <strong>#%d - %s</strong> has changed status from <strong>%s</strong> to <strong style="color: #f59e0b;">%s</strong>.</p>
			<p><a href="http://localhost:8081/tickets/%s" style="background-color: #6366f1; color: #ffffff; padding: 10px 20px; text-decoration: none; border-radius: 8px; font-weight: bold;">View Details</a></p>
		</div>
	`, ticket.TicketNumber, ticket.Title, oldStatus, newStatus, ticket.ID)

	s.Enqueue(EmailJob{To: recipientEmail, Subject: subject, HTML: body})
}

func (s *EmailService) NotifyNewComment(ticket *models.Ticket, authorName, commentBody string, recipientEmail string) {
	subject := fmt.Sprintf("[ticDesk] New Reply on Ticket #%d", ticket.TicketNumber)
	body := fmt.Sprintf(`
		<div style="font-family: sans-serif; background-color: #0f172a; color: #f8fafc; padding: 24px; borderRadius: 12px;">
			<h2 style="color: #6366f1;">New Reply Received</h2>
			<p><strong>%s</strong> commented on ticket <strong>#%d (%s)</strong>:</p>
			<div style="background-color: #1e293b; padding: 16px; border-radius: 8px; margin: 16px 0; font-style: italic;">
				"%s"
			</div>
			<p><a href="http://localhost:8081/tickets/%s" style="background-color: #6366f1; color: #ffffff; padding: 10px 20px; text-decoration: none; border-radius: 8px; font-weight: bold;">Reply on ticDesk</a></p>
		</div>
	`, authorName, ticket.TicketNumber, ticket.Title, commentBody, ticket.ID)

	s.Enqueue(EmailJob{To: recipientEmail, Subject: subject, HTML: body})
}
