package repository

import (
	"context"
	"fmt"
	"ticDesk/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AttachmentRepository struct {
	db *pgxpool.Pool
}

func NewAttachmentRepository(db *pgxpool.Pool) *AttachmentRepository {
	return &AttachmentRepository{db: db}
}

func (r *AttachmentRepository) CreateAttachment(ctx context.Context, ticketID *string, commentID *string, uploadedBy, fileName, filePath string, fileSize int64, mimeType string) (*models.Attachment, error) {
	query := `
		INSERT INTO attachments (ticket_id, comment_id, uploaded_by, file_name, file_path, file_size, mime_type)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, ticket_id, comment_id, uploaded_by, file_name, file_path, file_size, mime_type, created_at
	`
	a := &models.Attachment{}
	err := r.db.QueryRow(ctx, query, ticketID, commentID, uploadedBy, fileName, filePath, fileSize, mimeType).Scan(
		&a.ID, &a.TicketID, &a.CommentID, &a.UploadedBy, &a.FileName, &a.FilePath, &a.FileSize, &a.MimeType, &a.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to save attachment: %w", err)
	}
	return a, nil
}

func (r *AttachmentRepository) GetAttachmentByID(ctx context.Context, id string) (*models.Attachment, error) {
	query := `
		SELECT id, ticket_id, comment_id, uploaded_by, file_name, file_path, file_size, mime_type, created_at
		FROM attachments
		WHERE id = $1
	`
	a := &models.Attachment{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&a.ID, &a.TicketID, &a.CommentID, &a.UploadedBy, &a.FileName, &a.FilePath, &a.FileSize, &a.MimeType, &a.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return a, nil
}
