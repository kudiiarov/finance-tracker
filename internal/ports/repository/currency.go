package repository

import (
	"context"

	"github.com/sereozha/finance-tracker/internal/domain"
)

// CurrencyRepository manages configurable currencies.
type CurrencyRepository interface {
	// Create adds a new currency to the registry.
	Create(ctx context.Context, currency *domain.CurrencyInfo) error

	// Get retrieves a currency by code.
	Get(ctx context.Context, code string) (*domain.CurrencyInfo, error)

	// Update updates currency details.
	Update(ctx context.Context, currency *domain.CurrencyInfo) error

	// UpdateProviderPrefs updates the provider preferences for a currency.
	UpdateProviderPrefs(ctx context.Context, code string, prefs []domain.ProviderPreference) error

	// ListActive returns all active currencies.
	ListActive(ctx context.Context) ([]domain.CurrencyInfo, error)

	// ListByType returns currencies filtered by type.
	ListByType(ctx context.Context, currencyType domain.CurrencyType) ([]domain.CurrencyInfo, error)

	// Delete marks a currency as inactive (soft delete).
	Delete(ctx context.Context, code string) error
}
