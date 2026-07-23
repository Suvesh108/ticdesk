package services

import (
	"context"
	"fmt"
	"log"
	"net/smtp"
	"strings"
	"ticDesk/internal/models"
	"ticDesk/internal/repository"
)

type EmailJob struct {
	To      string
	Subject string
	HTML    string
}

type EmailService struct {
	smtpHost    string
	smtpPort    string
	fromAddr    string
	jobChan     chan EmailJob
	ticmailRepo *repository.TicMailRepository
}

func NewEmailService(host, port, from string, ticmailRepo *repository.TicMailRepository) *EmailService {
	return &EmailService{
		smtpHost:    host,
		smtpPort:    port,
		fromAddr:    from,
		jobChan:     make(chan EmailJob, 100),
		ticmailRepo: ticmailRepo,
	}
}

func (s *EmailService) StartWorker(ctx context.Context) {
	log.Printf("ticMail built-in email worker active. Listening on channel...")
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("ticMail worker stopping...")
				return
			case job, ok := <-s.jobChan:
				if !ok {
					return
				}
				// 1. Log to built-in ticMail system
				if s.ticmailRepo != nil {
					_, _ = s.ticmailRepo.LogEmail(ctx, job.To, job.Subject, job.HTML, "DELIVERED")
				}

				// 2. Optional SMTP dispatch (MailHog or real production SMTP)
				err := s.sendSMTP(job)
				if err != nil {
					log.Printf("ticMail captured alert for %s [%s] (SMTP bypass active)", job.To, job.Subject)
				} else {
					log.Printf("ticMail successfully dispatched SMTP alert for %s [%s]", job.To, job.Subject)
				}
			}
		}
	}()
}

func (s *EmailService) Enqueue(job EmailJob) {
	select {
	case s.jobChan <- job:
	default:
		log.Printf("Warning: ticMail queue full, dropping alert for %s", job.To)
	}
}

func (s *EmailService) sendSMTP(job EmailJob) error {
	addr := fmt.Sprintf("%s:%s", s.smtpHost, s.smtpPort)

	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("ticDesk ticMail <%s>", s.fromAddr)
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

	return smtp.SendMail(addr, nil, s.fromAddr, []string{job.To}, []byte(msg.String()))
}

// Outlook-Style Email Helper Templates

func (s *EmailService) buildOutlookWrapper(title, preheader, bodyContent, actionURL, actionText string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
    <style>
        body { font-family: 'Segoe UI', -apple-system, BlinkMacSystemFont, Roboto, Helvetica, Arial, sans-serif; background-color: #f3f2f1; color: #201f1e; margin: 0; padding: 0; }
        .outlook-container { max-width: 620px; margin: 24px auto; background-color: #ffffff; border-radius: 4px; border: 1px solid #e1dfdd; box-shadow: 0 1.6px 3.6px 0 rgba(0,0,0,0.132), 0 0.3px 0.9px 0 rgba(0,0,0,0.108); overflow: hidden; }
        .outlook-header { background-color: #0078d4; color: #ffffff; padding: 18px 24px; display: flex; align-items: center; justify-content: space-between; }
        .outlook-brand { font-size: 18px; font-weight: 700; letter-spacing: -0.2px; }
        .outlook-sub { font-size: 11px; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.5px; }
        .outlook-body { padding: 28px 24px; }
        .outlook-title { font-size: 20px; font-weight: 600; color: #106ebe; margin-top: 0; margin-bottom: 8px; }
        .outlook-pre { font-size: 13px; color: #605e5c; margin-bottom: 20px; }
        .outlook-table { width: 100%%; border-collapse: collapse; margin: 16px 0; background-color: #faf9f8; border: 1px solid #edebe9; border-radius: 4px; }
        .outlook-table td { padding: 10px 14px; font-size: 13px; border-bottom: 1px solid #edebe9; }
        .outlook-label { font-weight: 600; color: #323130; width: 130px; }
        .outlook-badge-open { display: inline-block; padding: 2px 8px; background-color: #deecf9; color: #005a9e; font-weight: 600; font-size: 11px; border-radius: 2px; }
        .outlook-badge-progress { display: inline-block; padding: 2px 8px; background-color: #fff4ce; color: #797775; font-weight: 600; font-size: 11px; border-radius: 2px; }
        .outlook-badge-resolved { display: inline-block; padding: 2px 8px; background-color: #dff6dd; color: #107c41; font-weight: 600; font-size: 11px; border-radius: 2px; }
        .outlook-btn { display: inline-block; background-color: #0078d4; color: #ffffff !important; font-size: 13px; font-weight: 600; text-decoration: none; padding: 10px 22px; border-radius: 2px; margin-top: 18px; }
        .outlook-footer { background-color: #faf9f8; padding: 16px 24px; font-size: 11px; color: #a19f9d; border-top: 1px solid #edebe9; text-align: center; }
    </style>
</head>
<body>
    <div class="outlook-container">
        <!-- Outlook Brand Bar -->
        <div class="outlook-header">
            <div class="outlook-brand">ticMail Alert Engine</div>
            <div class="outlook-sub">Microsoft Outlook 365 Integration</div>
        </div>

        <!-- Email Main Content -->
        <div class="outlook-body">
            <h1 class="outlook-title">%s</h1>
            <p class="outlook-pre">%s</p>

            %s

            <a href="%s" class="outlook-btn">%s</a>
        </div>

        <!-- Footer -->
        <div class="outlook-footer">
            Sent via ticMail Built-in Email Engine • ticDesk Helpdesk Platform
        </div>
    </div>
</body>
</html>
	`, title, title, preheader, bodyContent, actionURL, actionText)
}

func (s *EmailService) NotifyTicketCreated(ticket *models.Ticket, recipientEmail string) {
	subject := fmt.Sprintf("[ticMail Alert] Ticket #%d: %s", ticket.TicketNumber, ticket.Title)
	preheader := fmt.Sprintf("A new support ticket has been created by %s", recipientEmail)

	bodyTable := fmt.Sprintf(`
		<table class="outlook-table">
			<tr>
				<td class="outlook-label">Ticket Ref #:</td>
				<td><strong>#%d</strong></td>
			</tr>
			<tr>
				<td class="outlook-label">Subject:</td>
				<td>%s</td>
			</tr>
			<tr>
				<td class="outlook-label">Priority:</td>
				<td>%s</td>
			</tr>
			<tr>
				<td class="outlook-label">Status:</td>
				<td><span class="outlook-badge-open">OPEN</span></td>
			</tr>
		</table>
	`, ticket.TicketNumber, ticket.Title, ticket.Priority)

	actionURL := fmt.Sprintf("http://localhost:8081/tickets/%s", ticket.ID)
	html := s.buildOutlookWrapper("Ticket Created Successfully", preheader, bodyTable, actionURL, "Open Ticket in ticDesk")

	s.Enqueue(EmailJob{To: recipientEmail, Subject: subject, HTML: html})
}

func (s *EmailService) NotifyStatusChanged(ticket *models.Ticket, oldStatus, newStatus models.TicketStatus, recipientEmail string) {
	subject := fmt.Sprintf("[ticMail Update] Ticket #%d changed to %s", ticket.TicketNumber, newStatus)
	preheader := fmt.Sprintf("Ticket #%d status updated from %s to %s", ticket.TicketNumber, oldStatus, newStatus)

	badgeClass := "outlook-badge-progress"
	if newStatus == models.StatusResolved || newStatus == models.StatusClosed {
		badgeClass = "outlook-badge-resolved"
	}

	bodyTable := fmt.Sprintf(`
		<table class="outlook-table">
			<tr>
				<td class="outlook-label">Ticket Ref #:</td>
				<td><strong>#%d</strong></td>
			</tr>
			<tr>
				<td class="outlook-label">Previous Status:</td>
				<td>%s</td>
			</tr>
			<tr>
				<td class="outlook-label">New Status:</td>
				<td><span class="%s">%s</span></td>
			</tr>
		</table>
	`, ticket.TicketNumber, oldStatus, badgeClass, strings.ToUpper(string(newStatus)))

	actionURL := fmt.Sprintf("http://localhost:8081/tickets/%s", ticket.ID)
	html := s.buildOutlookWrapper("Ticket Status Updated", preheader, bodyTable, actionURL, "View Ticket Status")

	s.Enqueue(EmailJob{To: recipientEmail, Subject: subject, HTML: html})
}

func (s *EmailService) NotifyNewComment(ticket *models.Ticket, authorName, commentBody string, recipientEmail string) {
	subject := fmt.Sprintf("[ticMail Alert] New Reply on Ticket #%d", ticket.TicketNumber)
	preheader := fmt.Sprintf("%s replied to ticket #%d", authorName, ticket.TicketNumber)

	bodyTable := fmt.Sprintf(`
		<div style="background-color: #faf9f8; border-left: 4px solid #0078d4; padding: 14px 16px; margin: 16px 0; border-radius: 2px;">
			<div style="font-size: 12px; font-weight: 600; color: #323130; margin-bottom: 6px;">From: %s</div>
			<div style="font-size: 13px; color: #201f1e; line-height: 1.5;">%s</div>
		</div>
	`, authorName, commentBody)

	actionURL := fmt.Sprintf("http://localhost:8081/tickets/%s", ticket.ID)
	html := s.buildOutlookWrapper("New Message Received", preheader, bodyTable, actionURL, "Reply via ticDesk")

	s.Enqueue(EmailJob{To: recipientEmail, Subject: subject, HTML: html})
}
