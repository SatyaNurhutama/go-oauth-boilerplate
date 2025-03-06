package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/satya-nurhutama/go-oauth-boilerplate/internal/auth/entity"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByEmail(email string) (*entity.User, error) {
	query := `
		SELECT id, email, password, name, provider, provider_id
		FROM users
		WHERE email = $1
	`
	user := &entity.User{}
	err := r.db.QueryRowContext(context.Background(), query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Name,
		&user.Provider,
		&user.ProviderID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	return user, nil
}

func (r *UserRepository) FindByProvider(provider, providerID string) (*entity.User, error) {
	query := `
		SELECT id, email, name, provider, provider_id
		FROM users
		WHERE provider = $1 AND provider_id = $2
	`
	user := &entity.User{}
	err := r.db.QueryRowContext(context.Background(), query, provider, providerID).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Provider,
		&user.ProviderID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	return user, nil
}

func (r *UserRepository) Create(user *entity.User) error {
	query := `
		INSERT INTO users (email, password, name, provider, provider_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	err := r.db.QueryRowContext(
		context.Background(),
		query,
		user.Email,
		user.Password,
		user.Name,
		user.Provider,
		user.ProviderID,
	).Scan(&user.ID)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *UserRepository) FindOrCreateUserByProvider(provider, email, providerID, name string) (*entity.User, error) {
	// Check if the user already exists
	query := `
		SELECT id, email, name, provider, provider_id
		FROM users
		WHERE provider = $1 AND provider_id = $2
	`
	user := &entity.User{}
	err := r.db.QueryRowContext(context.Background(), query, provider, providerID).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Provider,
		&user.ProviderID,
	)
	if err == nil {
		return user, nil
	}

	// If the user doesn't exist, create a new user
	query = `
		INSERT INTO users (email, name, provider, provider_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`
	err = r.db.QueryRowContext(context.Background(), query, email, name, provider, providerID).Scan(&user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	user.Email = email
	user.Name = name
	user.Provider = provider
	user.ProviderID = providerID

	return user, nil
}
