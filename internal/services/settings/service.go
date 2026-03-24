// Package settings provides user settings business logic.
package settings

import (
	"context"
	"fmt"

	"github.com/sereozha/finance-tracker/internal/domain"
	"github.com/sereozha/finance-tracker/internal/ports/repository"
	"github.com/sereozha/finance-tracker/internal/ports/service"
)

// DefaultMainCurrency is the fallback currency.
const DefaultMainCurrency = "USD"

// Service implements service.SettingsService.
type Service struct {
	repo repository.SettingsRepository
}

// NewService creates a new settings service.
func NewService(repo repository.SettingsRepository) service.SettingsService {
	return &Service{repo: repo}
}

// Get retrieves settings for a user.
func (s *Service) Get(ctx context.Context, userID string) (*domain.Settings, error) {
	return s.repo.Get(ctx, userID)
}

// Update updates all user settings (full replacement).
func (s *Service) Update(ctx context.Context, userID string, mainCurrency string) error {
	if _, ok := domain.GetCurrency(mainCurrency); !ok {
		return fmt.Errorf("invalid currency: %s", mainCurrency)
	}
	return s.repo.UpdateMainCurrency(ctx, userID, mainCurrency)
}

// UpdateMainCurrency changes the user's main currency.
func (s *Service) UpdateMainCurrency(ctx context.Context, userID string, currency string) error {
	if _, ok := domain.GetCurrency(currency); !ok {
		return fmt.Errorf("invalid currency: %s", currency)
	}
	return s.repo.UpdateMainCurrency(ctx, userID, currency)
}

// GetMainCurrency returns the user's main currency, defaulting to USD.
func (s *Service) GetMainCurrency(ctx context.Context, userID string) (string, error) {
	settings, err := s.repo.Get(ctx, userID)
	if err != nil {
		return DefaultMainCurrency, nil
	}
	if settings.MainCurrency == "" {
		return DefaultMainCurrency, nil
	}
	return settings.MainCurrency, nil
}
