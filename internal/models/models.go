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

type User struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         UserRole  `json:"role"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}
