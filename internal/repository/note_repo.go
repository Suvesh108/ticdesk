package repository

import (
	"context"
	"ticDesk/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type NoteRepository struct {
	db *pgxpool.Pool
}

func NewNoteRepository(db *pgxpool.Pool) *NoteRepository {
	return &NoteRepository{db: db}
}

func (r *NoteRepository) CreateNote(ctx context.Context, userID, title, content, color string, isPinned bool) (*models.UserNote, error) {
	query := `
		INSERT INTO user_notes (user_id, title, content, color, is_pinned)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, title, content, color, is_pinned, created_at, updated_at;
	`
	var n models.UserNote
	err := r.db.QueryRow(ctx, query, userID, title, content, color, isPinned).Scan(
		&n.ID, &n.UserID, &n.Title, &n.Content, &n.Color, &n.IsPinned, &n.CreatedAt, &n.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *NoteRepository) ListNotes(ctx context.Context, userID string) ([]models.UserNote, error) {
	query := `
		SELECT id, user_id, title, content, color, is_pinned, created_at, updated_at
		FROM user_notes
		WHERE user_id = $1
		ORDER BY is_pinned DESC, updated_at DESC;
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []models.UserNote
	for rows.Next() {
		var n models.UserNote
		if err := rows.Scan(&n.ID, &n.UserID, &n.Title, &n.Content, &n.Color, &n.IsPinned, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}
	return notes, nil
}

func (r *NoteRepository) TogglePin(ctx context.Context, id, userID string) error {
	_, err := r.db.Exec(ctx, `UPDATE user_notes SET is_pinned = NOT is_pinned WHERE id = $1 AND user_id = $2`, id, userID)
	return err
}

func (r *NoteRepository) DeleteNote(ctx context.Context, id, userID string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM user_notes WHERE id = $1 AND user_id = $2`, id, userID)
	return err
}
