package repository

import (
	"context"
	"fmt"
	"ticDesk/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(ctx context.Context, name, email, passwordHash string, role models.UserRole) (*models.User, error) {
	query := `
		INSERT INTO users (name, email, password_hash, role)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, email, role, is_active, created_at
	`
	user := &models.User{}
	err := r.db.QueryRow(ctx, query, name, email, passwordHash, role).Scan(
		&user.ID, &user.Name, &user.Email, &user.Role, &user.IsActive, &user.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	return user, nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, name, email, password_hash, role, is_active, created_at
		FROM users
		WHERE email = $1 AND is_active = true
	`
	user := &models.User{}
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.Role, &user.IsActive, &user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	query := `
		SELECT id, name, email, role, is_active, created_at
		FROM users
		WHERE id = $1 AND is_active = true
	`
	user := &models.User{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Name, &user.Email, &user.Role, &user.IsActive, &user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}
