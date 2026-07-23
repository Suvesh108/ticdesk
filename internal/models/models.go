package models

import (
	"time"
)

type UserRole string

const (
	RoleAdmin    UserRole = "admin"
	RoleSupport  UserRole = "support"
	RoleCustomer UserRole = "customer"
)

type TicketPriority string

const (
	PriorityLow    TicketPriority = "low"
	PriorityMedium TicketPriority = "medium"
	PriorityHigh   TicketPriority = "high"
)

type TicketStatus string

const (
	StatusOpen       TicketStatus = "open"
	StatusInProgress TicketStatus = "in_progress"
	StatusResolved   TicketStatus = "resolved"
	StatusClosed     TicketStatus = "closed"
)

type User struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         UserRole  `json:"role"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}

type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Ticket struct {
	ID             string         `json:"id"`
	TicketNumber   int            `json:"ticket_number"`
	Title          string         `json:"title"`
	Description    string         `json:"description"`
	CategoryID     *int           `json:"category_id"`
	CategoryName   string         `json:"category_name,omitempty"`
	Priority       TicketPriority `json:"priority"`
	Status         TicketStatus   `json:"status"`
	CreatedByID    string         `json:"created_by_id"`
	CreatedByName  string         `json:"created_by_name,omitempty"`
	AssignedToID   *string        `json:"assigned_to_id,omitempty"`
	AssignedToName *string        `json:"assigned_to_name,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	ResolvedAt     *time.Time     `json:"resolved_at,omitempty"`
}

type TicketStatusHistory struct {
	ID        string        `json:"id"`
	TicketID  string        `json:"ticket_id"`
	ChangedBy string        `json:"changed_by"`
	OldStatus *TicketStatus `json:"old_status,omitempty"`
	NewStatus TicketStatus  `json:"new_status"`
	ChangedAt time.Time     `json:"changed_at"`
}

type Comment struct {
	ID          string       `json:"id"`
	TicketID    string       `json:"ticket_id"`
	AuthorID    string       `json:"author_id"`
	AuthorName  string       `json:"author_name"`
	AuthorRole  UserRole     `json:"author_role"`
	Body        string       `json:"body"`
	IsInternal  bool         `json:"is_internal"`
	CreatedAt   time.Time    `json:"created_at"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

type Attachment struct {
	ID         string    `json:"id"`
	TicketID   *string   `json:"ticket_id,omitempty"`
	CommentID  *string   `json:"comment_id,omitempty"`
	UploadedBy string    `json:"uploaded_by"`
	FileName   string    `json:"file_name"`
	FilePath   string    `json:"file_path"`
	FileSize   int64     `json:"file_size"`
	MimeType   string    `json:"mime_type"`
	CreatedAt  time.Time `json:"created_at"`
}

type CategoryStat struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type AgentStat struct {
	AgentID   string `json:"agent_id"`
	AgentName string `json:"agent_name"`
	OpenCount int    `json:"open_count"`
}

type DashboardStats struct {
	OpenCount            int            `json:"open_count"`
	InProgressCount      int            `json:"in_progress_count"`
	ResolvedCount        int            `json:"resolved_count"`
	ClosedCount          int            `json:"closed_count"`
	LowPriorityCount     int            `json:"low_priority_count"`
	MediumPriorityCount   int            `json:"medium_priority_count"`
	HighPriorityCount    int            `json:"high_priority_count"`
	AvgResolutionTime    string         `json:"avg_resolution_time"`
	CategoryDistribution []CategoryStat `json:"category_distribution"`
	AgentWorkload        []AgentStat    `json:"agent_workload"`
}

// Outlook Calendar Schedule Event
type ScheduleEvent struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	EventType     string    `json:"event_type"` // 'maintenance', 'shift', 'deadline'
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	CreatedByID   string    `json:"created_by_id"`
	CreatedByName string    `json:"created_by_name,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

// Outlook Personal & Team Scratchpad Notes
type UserNote struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Color     string    `json:"color"` // 'blue', 'amber', 'emerald', 'purple'
	IsPinned  bool      `json:"is_pinned"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
