package service

import (
	"context"

	"github.com/sereozha/finance-tracker/internal/domain"
)

// ExchangeRateService defines exchange rate operations.
type ExchangeRateService interface {
	// FetchRates fetches all currency rates from external providers and stores them.
	FetchRates(ctx context.Context) error

	// GetRate returns the exchange rate between two currencies.
	// Uses cached data. Calculates cross-rate via USD if needed.
	GetRate(ctx context.Context, from, to string) (*domain.ExchangeRate, error)

	// GetRateAt returns the historical rate at a specific time.
	// at is Unix timestamp.
	GetRateAt(ctx context.Context, from, to string, at int64) (*domain.ExchangeRate, error)
}
