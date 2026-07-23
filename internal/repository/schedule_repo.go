package repository

import (
	"context"
	"ticDesk/internal/models"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ScheduleRepository struct {
	db *pgxpool.Pool
}

func NewScheduleRepository(db *pgxpool.Pool) *ScheduleRepository {
	return &ScheduleRepository{db: db}
}

func (r *ScheduleRepository) CreateEvent(ctx context.Context, title, description, eventType string, startTime, endTime time.Time, createdByID string) (*models.ScheduleEvent, error) {
	query := `
		INSERT INTO schedule_events (title, description, event_type, start_time, end_time, created_by)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, title, description, event_type, start_time, end_time, created_by, created_at;
	`
	var e models.ScheduleEvent
	err := r.db.QueryRow(ctx, query, title, description, eventType, startTime, endTime, createdByID).Scan(
		&e.ID, &e.Title, &e.Description, &e.EventType, &e.StartTime, &e.EndTime, &e.CreatedByID, &e.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *ScheduleRepository) ListEvents(ctx context.Context) ([]models.ScheduleEvent, error) {
	query := `
		SELECT e.id, e.title, COALESCE(e.description, ''), e.event_type, e.start_time, e.end_time, e.created_by, u.name, e.created_at
		FROM schedule_events e
		JOIN users u ON e.created_by = u.id
		ORDER BY e.start_time ASC;
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.ScheduleEvent
	for rows.Next() {
		var e models.ScheduleEvent
		if err := rows.Scan(&e.ID, &e.Title, &e.Description, &e.EventType, &e.StartTime, &e.EndTime, &e.CreatedByID, &e.CreatedByName, &e.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}

func (r *ScheduleRepository) DeleteEvent(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM schedule_events WHERE id = $1`, id)
	return err
}
