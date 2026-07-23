package repository

import (
	"context"
	"fmt"
	"ticDesk/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TicketRepository struct {
	db *pgxpool.Pool
}

func NewTicketRepository(db *pgxpool.Pool) *TicketRepository {
	return &TicketRepository{db: db}
}

func (r *TicketRepository) GetCategories(ctx context.Context) ([]models.Category, error) {
	query := `SELECT id, name FROM categories ORDER BY name ASC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch categories: %w", err)
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, nil
}

func (r *TicketRepository) CreateTicket(ctx context.Context, title, description string, categoryID int, priority models.TicketPriority, createdByID string) (*models.Ticket, error) {
	query := `
		INSERT INTO tickets (title, description, category_id, priority, created_by)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, ticket_number, title, description, category_id, priority, status, created_by, created_at, updated_at
	`
	t := &models.Ticket{}
	err := r.db.QueryRow(ctx, query, title, description, categoryID, priority, createdByID).Scan(
		&t.ID, &t.TicketNumber, &t.Title, &t.Description, &t.CategoryID, &t.Priority, &t.Status, &t.CreatedByID, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create ticket: %w", err)
	}

	// Create initial status history entry
	historyQuery := `INSERT INTO ticket_status_history (ticket_id, changed_by, new_status) VALUES ($1, $2, $3)`
	_, _ = r.db.Exec(ctx, historyQuery, t.ID, createdByID, t.Status)

	return t, nil
}

func (r *TicketRepository) GetTicketByID(ctx context.Context, id string) (*models.Ticket, error) {
	query := `
		SELECT 
			t.id, t.ticket_number, t.title, t.description, t.category_id, 
			COALESCE(c.name, '') as category_name,
			t.priority, t.status, 
			t.created_by, u1.name as created_by_name,
			t.assigned_to, COALESCE(u2.name, '') as assigned_to_name,
			t.created_at, t.updated_at, t.resolved_at
		FROM tickets t
		LEFT JOIN categories c ON t.category_id = c.id
		INNER JOIN users u1 ON t.created_by = u1.id
		LEFT JOIN users u2 ON t.assigned_to = u2.id
		WHERE t.id = $1
	`
	t := &models.Ticket{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&t.ID, &t.TicketNumber, &t.Title, &t.Description, &t.CategoryID,
		&t.CategoryName, &t.Priority, &t.Status,
		&t.CreatedByID, &t.CreatedByName,
		&t.AssignedToID, &t.AssignedToName,
		&t.CreatedAt, &t.UpdatedAt, &t.ResolvedAt,
	)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *TicketRepository) ListTickets(ctx context.Context, user *models.User) ([]models.Ticket, error) {
	query := `
		SELECT 
			t.id, t.ticket_number, t.title, t.description, t.category_id, 
			COALESCE(c.name, '') as category_name,
			t.priority, t.status, 
			t.created_by, u1.name as created_by_name,
			t.assigned_to, COALESCE(u2.name, '') as assigned_to_name,
			t.created_at, t.updated_at, t.resolved_at
		FROM tickets t
		LEFT JOIN categories c ON t.category_id = c.id
		INNER JOIN users u1 ON t.created_by = u1.id
		LEFT JOIN users u2 ON t.assigned_to = u2.id
	`

	var args []interface{}
	// Customers see only their own tickets
	if user.Role == models.RoleCustomer {
		query += ` WHERE t.created_by = $1 `
		args = append(args, user.ID)
	}

	query += ` ORDER BY t.created_at DESC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list tickets: %w", err)
	}
	defer rows.Close()

	var tickets []models.Ticket
	for rows.Next() {
		var t models.Ticket
		if err := rows.Scan(
			&t.ID, &t.TicketNumber, &t.Title, &t.Description, &t.CategoryID,
			&t.CategoryName, &t.Priority, &t.Status,
			&t.CreatedByID, &t.CreatedByName,
			&t.AssignedToID, &t.AssignedToName,
			&t.CreatedAt, &t.UpdatedAt, &t.ResolvedAt,
		); err != nil {
			return nil, err
		}
		tickets = append(tickets, t)
	}
	return tickets, nil
}
