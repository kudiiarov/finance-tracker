package repository

import (
	"context"

	"github.com/sereozha/finance-tracker/internal/domain"
)

// ExchangeRateRepository defines storage operations for exchange rates.
type ExchangeRateRepository interface {
	// Save stores a new exchange rate.
	Save(ctx context.Context, rate *domain.ExchangeRate) error

	// GetLatest returns the most recent rate for a currency pair.
	GetLatest(ctx context.Context, fromCurrency, toCurrency string) (*domain.ExchangeRate, error)

	// GetRateAt returns the rate closest to a specific time (for historical data).
	// at is Unix timestamp.
	GetRateAt(ctx context.Context, fromCurrency, toCurrency string, at int64) (*domain.ExchangeRate, error)

	// GetHistory returns rates within a time range.
	// from and to are Unix timestamps.
	GetHistory(ctx context.Context, fromCurrency, toCurrency string, from, to int64) ([]domain.ExchangeRate, error)

	// DeleteOlderThan removes rates older than the specified time.
	// before is Unix timestamp.
	DeleteOlderThan(ctx context.Context, before int64) error
}
