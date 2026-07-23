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
	ID           string         `json:"id"`
	TicketNumber int            `json:"ticket_number"`
	Title        string         `json:"title"`
	Description  string         `json:"description"`
	CategoryID   *int           `json:"category_id"`
	CategoryName string         `json:"category_name,omitempty"`
	Priority     TicketPriority `json:"priority"`
	Status       TicketStatus   `json:"status"`
	CreatedByID  string         `json:"created_by_id"`
	CreatedByName string        `json:"created_by_name,omitempty"`
	AssignedToID *string        `json:"assigned_to_id,omitempty"`
	AssignedToName *string      `json:"assigned_to_name,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	ResolvedAt   *time.Time     `json:"resolved_at,omitempty"`
}

type TicketStatusHistory struct {
	ID        string       `json:"id"`
	TicketID  string       `json:"ticket_id"`
	ChangedBy string       `json:"changed_by"`
	OldStatus *TicketStatus`json:"old_status,omitempty"`
	NewStatus TicketStatus `json:"new_status"`
	ChangedAt time.Time    `json:"changed_at"`
}
