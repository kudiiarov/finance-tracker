package service

import (
	"context"

	"github.com/sereozha/finance-tracker/internal/domain"
)

// SettingsService defines user settings operations.
type SettingsService interface {
	Get(ctx context.Context, userID string) (*domain.Settings, error)
	Update(ctx context.Context, userID string, mainCurrency string) error
	UpdateMainCurrency(ctx context.Context, userID string, currency string) error
	GetMainCurrency(ctx context.Context, userID string) (string, error)
}
