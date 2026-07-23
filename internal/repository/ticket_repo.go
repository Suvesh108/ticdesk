package repository

import (
	"context"
	"fmt"
	"ticDesk/internal/models"
	"time"

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

func (r *TicketRepository) GetSupportAgents(ctx context.Context) ([]models.User, error) {
	query := `SELECT id, name, email, role FROM users WHERE role IN ('admin', 'support') AND is_active = true ORDER BY name ASC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch agents: %w", err)
	}
	defer rows.Close()

	var agents []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Role); err != nil {
			return nil, err
		}
		agents = append(agents, u)
	}
	return agents, nil
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

func (r *TicketRepository) UpdateTicketStatus(ctx context.Context, ticketID string, newStatus models.TicketStatus, changedBy string) (*models.Ticket, error) {
	// Fetch current status first
	currentTicket, err := r.GetTicketByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	var resolvedAt *time.Time
	if newStatus == models.StatusResolved || newStatus == models.StatusClosed {
		now := time.Now()
		resolvedAt = &now
	}

	query := `
		UPDATE tickets 
		SET status = $1, updated_at = now(), resolved_at = COALESCE($2, resolved_at)
		WHERE id = $3
	`
	_, err = r.db.Exec(ctx, query, newStatus, resolvedAt, ticketID)
	if err != nil {
		return nil, fmt.Errorf("failed to update ticket status: %w", err)
	}

	// Insert into status history
	historyQuery := `INSERT INTO ticket_status_history (ticket_id, changed_by, old_status, new_status) VALUES ($1, $2, $3, $4)`
	_, _ = r.db.Exec(ctx, historyQuery, ticketID, changedBy, currentTicket.Status, newStatus)

	return r.GetTicketByID(ctx, ticketID)
}

func (r *TicketRepository) UpdateTicketPriority(ctx context.Context, ticketID string, newPriority models.TicketPriority) (*models.Ticket, error) {
	query := `UPDATE tickets SET priority = $1, updated_at = now() WHERE id = $2`
	_, err := r.db.Exec(ctx, query, newPriority, ticketID)
	if err != nil {
		return nil, fmt.Errorf("failed to update ticket priority: %w", err)
	}
	return r.GetTicketByID(ctx, ticketID)
}

func (r *TicketRepository) UpdateTicketAssignee(ctx context.Context, ticketID string, assigneeID *string) (*models.Ticket, error) {
	query := `UPDATE tickets SET assigned_to = $1, updated_at = now() WHERE id = $2`
	_, err := r.db.Exec(ctx, query, assigneeID, ticketID)
	if err != nil {
		return nil, fmt.Errorf("failed to update assignee: %w", err)
	}
	return r.GetTicketByID(ctx, ticketID)
}
