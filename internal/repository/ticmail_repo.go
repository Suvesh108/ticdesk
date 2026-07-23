package repository

import (
	"context"
	"ticDesk/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TicMailRepository struct {
	db *pgxpool.Pool
}

func NewTicMailRepository(db *pgxpool.Pool) *TicMailRepository {
	return &TicMailRepository{db: db}
}

func (r *TicMailRepository) LogEmail(ctx context.Context, recipient, subject, bodyHTML, status string) (*models.TicMailLog, error) {
	query := `
		INSERT INTO ticmail_logs (recipient, subject, body_html, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, recipient, subject, body_html, status, created_at;
	`
	var logItem models.TicMailLog
	err := r.db.QueryRow(ctx, query, recipient, subject, bodyHTML, status).Scan(
		&logItem.ID, &logItem.Recipient, &logItem.Subject, &logItem.BodyHTML, &logItem.Status, &logItem.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &logItem, nil
}

func (r *TicMailRepository) ListLogs(ctx context.Context) ([]models.TicMailLog, error) {
	query := `
		SELECT id, recipient, subject, body_html, status, created_at
		FROM ticmail_logs
		ORDER BY created_at DESC
		LIMIT 100;
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.TicMailLog
	for rows.Next() {
		var l models.TicMailLog
		if err := rows.Scan(&l.ID, &l.Recipient, &l.Subject, &l.BodyHTML, &l.Status, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}

func (r *TicMailRepository) ClearLogs(ctx context.Context) error {
	_, err := r.db.Exec(ctx, `TRUNCATE TABLE ticmail_logs`)
	return err
}
