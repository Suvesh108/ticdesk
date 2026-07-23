package repository

import (
	"context"
	"fmt"
	"ticDesk/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type CommentRepository struct {
	db *pgxpool.Pool
}

func NewCommentRepository(db *pgxpool.Pool) *CommentRepository {
	return &CommentRepository{db: db}
}

func (r *CommentRepository) CreateComment(ctx context.Context, ticketID, authorID, body string, isInternal bool) (*models.Comment, error) {
	query := `
		INSERT INTO comments (ticket_id, author_id, body, is_internal)
		VALUES ($1, $2, $3, $4)
		RETURNING id, ticket_id, author_id, body, is_internal, created_at
	`
	c := &models.Comment{}
	err := r.db.QueryRow(ctx, query, ticketID, authorID, body, isInternal).Scan(
		&c.ID, &c.TicketID, &c.AuthorID, &c.Body, &c.IsInternal, &c.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}

	// Fetch author details
	authorQuery := `SELECT name, role FROM users WHERE id = $1`
	_ = r.db.QueryRow(ctx, authorQuery, authorID).Scan(&c.AuthorName, &c.AuthorRole)

	return c, nil
}

func (r *CommentRepository) ListCommentsForTicket(ctx context.Context, ticketID string, userRole models.UserRole) ([]models.Comment, error) {
	query := `
		SELECT c.id, c.ticket_id, c.author_id, u.name as author_name, u.role as author_role, c.body, c.is_internal, c.created_at
		FROM comments c
		INNER JOIN users u ON c.author_id = u.id
		WHERE c.ticket_id = $1
	`

	// If customer, hide internal notes
	if userRole == models.RoleCustomer {
		query += ` AND c.is_internal = false `
	}

	query += ` ORDER BY c.created_at ASC`

	rows, err := r.db.Query(ctx, query, ticketID)
	if err != nil {
		return nil, fmt.Errorf("failed to list comments: %w", err)
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var c models.Comment
		if err := rows.Scan(
			&c.ID, &c.TicketID, &c.AuthorID, &c.AuthorName, &c.AuthorRole, &c.Body, &c.IsInternal, &c.CreatedAt,
		); err != nil {
			return nil, err
		}

		// Fetch attachments for this comment
		attQuery := `SELECT id, file_name, file_path, file_size, mime_type FROM attachments WHERE comment_id = $1`
		attRows, _ := r.db.Query(ctx, attQuery, c.ID)
		if attRows != nil {
			var atts []models.Attachment
			for attRows.Next() {
				var a models.Attachment
				_ = attRows.Scan(&a.ID, &a.FileName, &a.FilePath, &a.FileSize, &a.MimeType)
				atts = append(atts, a)
			}
			attRows.Close()
			c.Attachments = atts
		}

		comments = append(comments, c)
	}
	return comments, nil
}
