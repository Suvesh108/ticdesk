package repository

import (
	"context"
	"fmt"
	"ticDesk/internal/models"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TicketFilterOptions struct {
	Search     string
	CategoryID int
	Status     string
	Priority   string
	Page       int
	Limit      int
}

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
	tickets, _, err := r.ListTicketsFiltered(ctx, user, TicketFilterOptions{Page: 1, Limit: 100})
	return tickets, err
}

func (r *TicketRepository) ListTicketsFiltered(ctx context.Context, user *models.User, opts TicketFilterOptions) ([]models.Ticket, int, error) {
	if opts.Page < 1 {
		opts.Page = 1
	}
	if opts.Limit < 1 {
		opts.Limit = 10
	}
	offset := (opts.Page - 1) * opts.Limit

	whereConditions := []string{"1=1"}
	var args []interface{}
	argID := 1

	if user.Role == models.RoleCustomer {
		whereConditions = append(whereConditions, fmt.Sprintf("t.created_by = $%d", argID))
		args = append(args, user.ID)
		argID++
	}

	if opts.Search != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("(t.title ILIKE $%d OR t.description ILIKE $%d)", argID, argID))
		args = append(args, "%"+opts.Search+"%")
		argID++
	}

	if opts.CategoryID > 0 {
		whereConditions = append(whereConditions, fmt.Sprintf("t.category_id = $%d", argID))
		args = append(args, opts.CategoryID)
		argID++
	}

	if opts.Status != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("t.status = $%d", argID))
		args = append(args, opts.Status)
		argID++
	}

	if opts.Priority != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("t.priority = $%d", argID))
		args = append(args, opts.Priority)
		argID++
	}

	whereClause := " WHERE " + fmt.Sprintf("%s", whereConditions[0])
	for i := 1; i < len(whereConditions); i++ {
		whereClause += " AND " + whereConditions[i]
	}

	// Count total matching items for pagination
	countQuery := "SELECT COUNT(*) FROM tickets t " + whereClause
	var totalCount int
	_ = r.db.QueryRow(ctx, countQuery, args...).Scan(&totalCount)

	// Fetch page slice
	query := fmt.Sprintf(`
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
		%s
		ORDER BY t.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argID, argID+1)

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

	query := `
		UPDATE tickets 
		SET status = $1, updated_at = now(), resolved_at = COALESCE($2, resolved_at)
		WHERE id = $3
	`
	_, err = r.db.Exec(ctx, query, newStatus, resolvedAt, ticketID)
	if err != nil {
		return nil, fmt.Errorf("failed to update ticket status: %w", err)
	}

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
		GROUP BY c.name 
		ORDER BY count DESC
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

	// Agent Workload (Support & Admin)
	agentQuery := `
		SELECT u.id, u.name, COUNT(t.id) as open_count
		FROM users u
		LEFT JOIN tickets t ON u.id = t.assigned_to AND t.status IN ('open', 'in_progress')
		WHERE u.role IN ('admin', 'support') AND u.is_active = true
		GROUP BY u.id, u.name
		ORDER BY open_count DESC
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

	// Average resolution interval
	avgQuery := `
		SELECT AVG(resolved_at - created_at)
		FROM tickets
		WHERE resolved_at IS NOT NULL
	`
	var avgDuration *time.Duration
	if err := r.db.QueryRow(ctx, avgQuery).Scan(&avgDuration); err == nil && avgDuration != nil {
		stats.AvgResolutionTime = fmt.Sprintf("%dh %dm", int(avgDuration.Hours()), int(avgDuration.Minutes())%60)
	}

	return stats, nil
}
