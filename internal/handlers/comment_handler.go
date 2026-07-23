package handlers

import (
	"html/template"
	"net/http"
	"strings"
	"ticDesk/internal/auth"
	"ticDesk/internal/models"
	"ticDesk/internal/repository"
	"ticDesk/internal/services"

	"github.com/go-chi/chi/v5"
)

type CommentHandler struct {
	commentRepo    *repository.CommentRepository
	attachmentRepo *repository.AttachmentRepository
	ticketRepo     *repository.TicketRepository
	storageService services.StorageService
	emailService   *services.EmailService
}

func NewCommentHandler(
	commentRepo *repository.CommentRepository,
	attachmentRepo *repository.AttachmentRepository,
	ticketRepo *repository.TicketRepository,
	storageService services.StorageService,
	emailService *services.EmailService,
) *CommentHandler {
	return &CommentHandler{
		commentRepo:    commentRepo,
		attachmentRepo: attachmentRepo,
		ticketRepo:     ticketRepo,
		storageService: storageService,
		emailService:   emailService,
	}
}

type CommentThreadData struct {
	TicketID string
	Comments []models.Comment
	User     *models.User
}

func (h *CommentHandler) GetComments(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	ticketID := chi.URLParam(r, "id")

	comments, err := h.commentRepo.ListCommentsForTicket(r.Context(), ticketID, user.Role)
	if err != nil {
		http.Error(w, "Failed to load comments: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := CommentThreadData{
		TicketID: ticketID,
		Comments: comments,
		User:     user,
	}

	tmpl, err := template.ParseFiles("web/templates/partials/comment_list.html")
	if err != nil {
		http.Error(w, "Partial error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tmpl.Execute(w, data)
}

func (h *CommentHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	ticketID := chi.URLParam(r, "id")

	// Parse multipart form up to 10MB
	if err := r.ParseMultipartForm(10 * 1024 * 1024); err != nil {
		_ = r.ParseForm()
	}

	body := strings.TrimSpace(r.FormValue("body"))
	isInternalStr := r.FormValue("is_internal")
	isInternal := (isInternalStr == "true" || isInternalStr == "on")

	// Only admin and support can create internal notes
	if isInternal && user.Role != models.RoleAdmin && user.Role != models.RoleSupport {
		isInternal = false
	}

	if body == "" {
		http.Error(w, "Comment body cannot be empty", http.StatusBadRequest)
		return
	}

	comment, err := h.commentRepo.CreateComment(r.Context(), ticketID, user.ID, body, isInternal)
	if err != nil {
		http.Error(w, "Failed to post comment: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Process optional file upload attachment
	file, header, err := r.FormFile("file")
	if err == nil && file != nil {
		defer file.Close()
		filePath, fileSize, mimeType, err := h.storageService.SaveFile(ticketID, file, header)
		if err == nil {
			_, _ = h.attachmentRepo.CreateAttachment(r.Context(), &ticketID, &comment.ID, user.ID, header.Filename, filePath, fileSize, mimeType)
		}
	}

	// Trigger Email Notification for public comments (non-blocking)
	if !isInternal && h.emailService != nil {
		ticket, _ := h.ticketRepo.GetTicketByID(r.Context(), ticketID)
		if ticket != nil {
			h.emailService.NotifyNewComment(ticket, user.Name, body, user.Email)
		}
	}

	// Re-render complete comment list partial for HTMX swap
	comments, _ := h.commentRepo.ListCommentsForTicket(r.Context(), ticketID, user.Role)
	data := CommentThreadData{
		TicketID: ticketID,
		Comments: comments,
		User:     user,
	}

	tmpl, err := template.ParseFiles("web/templates/partials/comment_list.html")
	if err != nil {
		http.Error(w, "Partial error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tmpl.Execute(w, data)
}
