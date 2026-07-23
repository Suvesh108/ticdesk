package handlers

import (
	"net/http"
	"ticDesk/internal/auth"
	"ticDesk/internal/models"
	"ticDesk/internal/repository"

	"github.com/go-chi/chi/v5"
)

type AttachmentHandler struct {
	attachmentRepo *repository.AttachmentRepository
	ticketRepo     *repository.TicketRepository
}

func NewAttachmentHandler(
	attachmentRepo *repository.AttachmentRepository,
	ticketRepo *repository.TicketRepository,
) *AttachmentHandler {
	return &AttachmentHandler{
		attachmentRepo: attachmentRepo,
		ticketRepo:     ticketRepo,
	}
}

func (h *AttachmentHandler) DownloadAttachment(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	attachmentID := chi.URLParam(r, "id")

	att, err := h.attachmentRepo.GetAttachmentByID(r.Context(), attachmentID)
	if err != nil {
		http.Error(w, "Attachment not found", http.StatusNotFound)
		return
	}

	// RBAC Check: Ensure user has access to the ticket
	if att.TicketID != nil {
		ticket, err := h.ticketRepo.GetTicketByID(r.Context(), *att.TicketID)
		if err == nil && user.Role == models.RoleCustomer && ticket.CreatedByID != user.ID {
			http.Error(w, "Forbidden: you do not have permission to download this attachment", http.StatusForbidden)
			return
		}
	}

	w.Header().Set("Content-Disposition", "inline; filename=\""+att.FileName+"\"")
	w.Header().Set("Content-Type", att.MimeType)
	http.ServeFile(w, r, att.FilePath)
}
