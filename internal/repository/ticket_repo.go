package repository

import (
	"context"
	"fmt"
	"strings"
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

	// Run Intelligent Auto-Assignment Algorithm
	_, _ = r.AutoAssignTicket(ctx, t.ID)

	return r.GetTicketByID(ctx, t.ID)
}

func (r *TicketRepository) AutoAssignTicket(ctx context.Context, ticketID string) (*models.Ticket, error) {
	// Select active support agent with the lowest open ticket workload
	query := `
		SELECT u.id
		FROM users u
		LEFT JOIN tickets t ON u.id = t.assigned_to AND t.status IN ('open', 'in_progress')
		WHERE u.role IN ('admin', 'support') AND u.is_active = true
		GROUP BY u.id, u.created_at
		ORDER BY COUNT(t.id) ASC, u.created_at ASC
		LIMIT 1;
	`
	var selectedAgentID string
	err := r.db.QueryRow(ctx, query).Scan(&selectedAgentID)
	if err != nil {
		return nil, err
	}

	updateQuery := `UPDATE tickets SET assigned_to = $1, auto_assigned = true, updated_at = now() WHERE id = $2`
	_, err = r.db.Exec(ctx, updateQuery, selectedAgentID, ticketID)
	if err != nil {
		return nil, err
	}
	return r.GetTicketByID(ctx, ticketID)
}

func (r *TicketRepository) GetTicketByID(ctx context.Context, id string) (*models.Ticket, error) {
	query := `
		SELECT 
			t.id, t.ticket_number, t.title, t.description, t.category_id, 
			COALESCE(c.name, '') as category_name,
			t.priority, t.status, 
			t.created_by, COALESCE(u1.name, 'Guest Customer') as created_by_name,
			t.assigned_to, COALESCE(u2.name, '') as assigned_to_name,
			t.auto_assigned,
			t.created_at, t.updated_at, t.resolved_at
		FROM tickets t
		LEFT JOIN categories c ON t.category_id = c.id
		LEFT JOIN users u1 ON t.created_by = u1.id
		LEFT JOIN users u2 ON t.assigned_to = u2.id
		WHERE t.id = $1
	`
	t := &models.Ticket{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&t.ID, &t.TicketNumber, &t.Title, &t.Description, &t.CategoryID,
		&t.CategoryName, &t.Priority, &t.Status,
		&t.CreatedByID, &t.CreatedByName,
		&t.AssignedToID, &t.AssignedToName,
		&t.AutoAssigned,
		&t.CreatedAt, &t.UpdatedAt, &t.ResolvedAt,
	)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *TicketRepository) ListTickets(ctx context.Context, user *models.User) ([]models.Ticket, error) {
	tickets, _, err := r.ListTicketsFiltered(ctx, user, TicketFilterOptions{Page: 1, Limit: 100})
	return tickets, err
}

type TicketFilterOptions struct {
	Search     string
	Status     models.TicketStatus
	Priority   models.TicketPriority
	CategoryID int
	Page       int
	Limit      int
}

func (r *TicketRepository) ListTicketsFiltered(ctx context.Context, user *models.User, opts TicketFilterOptions) ([]models.Ticket, int, error) {
	if opts.Page <= 0 {
		opts.Page = 1
	}
	if opts.Limit <= 0 {
		opts.Limit = 10
	}
	offset := (opts.Page - 1) * opts.Limit

	whereClauses := []string{"1=1"}
	args := []interface{}{}
	argPos := 1

	if user.Role == models.RoleCustomer {
		whereClauses = append(whereClauses, fmt.Sprintf("t.created_by = $%d", argPos))
		args = append(args, user.ID)
		argPos++
	}

	if opts.Status != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("t.status = $%d", argPos))
		args = append(args, opts.Status)
		argPos++
	}

	if opts.Priority != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("t.priority = $%d", argPos))
		args = append(args, opts.Priority)
		argPos++
	}

	if opts.CategoryID > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("t.category_id = $%d", argPos))
		args = append(args, opts.CategoryID)
		argPos++
	}

	if opts.Search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("(t.title ILIKE $%d OR t.description ILIKE $%d)", argPos, argPos))
		args = append(args, "%"+opts.Search+"%")
		argPos++
	}

	whereSQL := ""
	if len(whereClauses) > 0 {
		whereSQL = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM tickets t %s", whereSQL)
	var totalCount int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count tickets: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT 
			t.id, t.ticket_number, t.title, t.description, t.category_id, 
			COALESCE(c.name, '') as category_name,
			t.priority, t.status, 
			t.created_by, COALESCE(u1.name, 'Guest Customer') as created_by_name,
			t.assigned_to, COALESCE(u2.name, '') as assigned_to_name,
			t.auto_assigned,
			t.created_at, t.updated_at, t.resolved_at
		FROM tickets t
		LEFT JOIN categories c ON t.category_id = c.id
		LEFT JOIN users u1 ON t.created_by = u1.id
		LEFT JOIN users u2 ON t.assigned_to = u2.id
		%s
		ORDER BY t.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, argPos, argPos+1)

	args = append(args, opts.Limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list tickets: %w", err)
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
			&t.AutoAssigned,
			&t.CreatedAt, &t.UpdatedAt, &t.ResolvedAt,
		); err != nil {
			return nil, 0, err
		}
		tickets = append(tickets, t)
	}

	return tickets, totalCount, nil
}

func (r *TicketRepository) UpdateTicketStatus(ctx context.Context, ticketID string, newStatus models.TicketStatus, changedBy string) (*models.Ticket, error) {
	currentTicket, err := r.GetTicketByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	var resolvedAt *time.Time
	if newStatus == models.StatusResolved || newStatus == models.StatusClosed {
		now := time.Now()
		resolvedAt = &now
	}

	query := `UPDATE tickets SET status = $1, resolved_at = $2, updated_at = now() WHERE id = $3`
	_, err = r.db.Exec(ctx, query, newStatus, resolvedAt, ticketID)
	if err != nil {
		return nil, fmt.Errorf("failed to update ticket status: %w", err)
	}

	historyQuery := `INSERT INTO ticket_status_history (ticket_id, changed_by, old_status, new_status) VALUES ($1, $2, $3, $4)`
	_, _ = r.db.Exec(ctx, historyQuery, ticketID, changedBy, currentTicket.Status, newStatus)

	// Automatic Temporary Guest Account Deactivation on Ticket Closure
	if newStatus == models.StatusClosed {
		deactivateGuestQuery := `
			UPDATE users 
			SET is_active = false 
			WHERE id = $1 AND is_temporary = true;
		`
		_, _ = r.db.Exec(ctx, deactivateGuestQuery, currentTicket.CreatedByID)
	}

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
	query := `UPDATE tickets SET assigned_to = $1, auto_assigned = false, updated_at = now() WHERE id = $2`
	_, err := r.db.Exec(ctx, query, assigneeID, ticketID)
	if err != nil {
		return nil, fmt.Errorf("failed to update assignee: %w", err)
	}
	return r.GetTicketByID(ctx, ticketID)
}

func (r *TicketRepository) GetDashboardStats(ctx context.Context, user *models.User) (*models.DashboardStats, error) {
	stats := &models.DashboardStats{
		AvgResolutionTime: "< 15m",
	}

	whereClause := ""
	var args []interface{}
	if user.Role == models.RoleCustomer {
		whereClause = " WHERE created_by = $1 "
		args = append(args, user.ID)
	}

	// Status counts
	statusQuery := fmt.Sprintf(`SELECT status, COUNT(*) FROM tickets %s GROUP BY status`, whereClause)
	rows, err := r.db.Query(ctx, statusQuery, args...)
	if err == nil {
		for rows.Next() {
			var st string
			var cnt int
			if err := rows.Scan(&st, &cnt); err == nil {
				switch models.TicketStatus(st) {
				case models.StatusOpen:
					stats.OpenCount = cnt
				case models.StatusInProgress:
					stats.InProgressCount = cnt
				case models.StatusResolved:
					stats.ResolvedCount = cnt
				case models.StatusClosed:
					stats.ClosedCount = cnt
				}
			}
		}
		rows.Close()
	}

	// Priority counts
	priorityQuery := fmt.Sprintf(`SELECT priority, COUNT(*) FROM tickets %s GROUP BY priority`, whereClause)
	pRows, err := r.db.Query(ctx, priorityQuery, args...)
	if err == nil {
		for pRows.Next() {
			var pr string
			var cnt int
			if err := pRows.Scan(&pr, &cnt); err == nil {
				switch models.TicketPriority(pr) {
				case models.PriorityLow:
					stats.LowPriorityCount = cnt
				case models.PriorityMedium:
					stats.MediumPriorityCount = cnt
				case models.PriorityHigh:
					stats.HighPriorityCount = cnt
				}
			}
		}
		pRows.Close()
	}

	// Category distribution
	catQuery := `
		SELECT c.name, COUNT(t.id) 
		FROM categories c 
		LEFT JOIN tickets t ON c.id = t.category_id 
		GROUP BY c.name ORDER BY c.name ASC
	`
	cRows, err := r.db.Query(ctx, catQuery)
	if err == nil {
		for cRows.Next() {
			var cs models.CategoryStat
			if err := cRows.Scan(&cs.Name, &cs.Count); err == nil {
				stats.CategoryDistribution = append(stats.CategoryDistribution, cs)
			}
		}
		cRows.Close()
	}

	// Agent Workload
	agentQuery := `
		SELECT u.id, u.name, COUNT(t.id) 
		FROM users u 
		LEFT JOIN tickets t ON u.id = t.assigned_to AND t.status IN ('open', 'in_progress')
		WHERE u.role IN ('admin', 'support') AND u.is_active = true
		GROUP BY u.id, u.name 
		ORDER BY u.name ASC
	`
	aRows, err := r.db.Query(ctx, agentQuery)
	if err == nil {
		for aRows.Next() {
			var ag models.AgentStat
			if err := aRows.Scan(&ag.AgentID, &ag.AgentName, &ag.OpenCount); err == nil {
				stats.AgentWorkload = append(stats.AgentWorkload, ag)
			}
		}
		aRows.Close()
	}

	return stats, nil
}
