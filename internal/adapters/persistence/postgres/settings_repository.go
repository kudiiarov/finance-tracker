package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sereozha/finance-tracker/internal/domain"
	"github.com/sereozha/finance-tracker/internal/ports/repository"
)

// SettingsRepository implements repository.SettingsRepository.
type SettingsRepository struct {
	pool *pgxpool.Pool
}

// NewSettingsRepository creates a new PostgreSQL repository.
func NewSettingsRepository(pool *pgxpool.Pool) repository.SettingsRepository {
	return &SettingsRepository{pool: pool}
}

// Get retrieves settings for a user.
func (r *SettingsRepository) Get(ctx context.Context, userID string) (*domain.Settings, error) {
	query := `
		SELECT user_id, main_currency, created_at, updated_at
		FROM user_settings
		WHERE user_id = $1
	`

	var s domain.Settings
	row := r.pool.QueryRow(ctx, query, userID)
	err := row.Scan(
		&s.UserID,
		&s.MainCurrency,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("settings not found for user: %s", userID)
		}
		return nil, fmt.Errorf("failed to get settings: %w", err)
	}

	return &s, nil
}

// Create creates settings for a new user.
func (r *SettingsRepository) Create(ctx context.Context, s *domain.Settings) error {
	query := `
		INSERT INTO user_settings (user_id, main_currency, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.pool.Exec(ctx, query,
		s.UserID,
		s.MainCurrency,
		s.CreatedAt,
		s.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create settings: %w", err)
	}

	return nil
}

// UpdateMainCurrency changes the user's main currency.
func (r *SettingsRepository) UpdateMainCurrency(ctx context.Context, userID string, currency string) error {
	query := `
		UPDATE user_settings
		SET main_currency = $1
		WHERE user_id = $2
	`

	result, err := r.pool.Exec(ctx, query, currency, userID)
	if err != nil {
		return fmt.Errorf("failed to update main currency: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("settings not found for user: %s", userID)
	}

	return nil
}
