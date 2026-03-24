package repository

import (
	"context"

	"github.com/sereozha/finance-tracker/internal/domain"
)

// SettingsRepository defines storage operations for user settings.
type SettingsRepository interface {
	// Get retrieves settings for a user.
	Get(ctx context.Context, userID string) (*domain.Settings, error)

	// Create creates settings for a new user.
	Create(ctx context.Context, settings *domain.Settings) error

	// UpdateMainCurrency changes the user's main currency.
	UpdateMainCurrency(ctx context.Context, userID string, currency string) error
}
