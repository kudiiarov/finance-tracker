package service

import (
	"context"

	"github.com/sereozha/finance-tracker/internal/domain"
)

// ConversionStrategy defines how to derive exchange rates.
type ConversionStrategy interface {
	// Name returns the strategy identifier.
	Name() string

	// GetRate returns the exchange rate between two currencies.
	// The strategy may use multiple stored rates to derive the result.
	GetRate(ctx context.Context, from, to string) (*domain.ExchangeRate, error)

	// CanConvert checks if this strategy can convert the given pair.
	CanConvert(ctx context.Context, from, to string) bool
}

// RateRepository defines the storage interface needed by conversion strategies.
type RateRepository interface {
	// GetLatest returns the most recent rate for a currency pair.
	GetLatest(ctx context.Context, fromCurrency, toCurrency string) (*domain.ExchangeRate, error)

	// GetRateAt returns the rate closest to a specific time.
	GetRateAt(ctx context.Context, fromCurrency, toCurrency string, at int64) (*domain.ExchangeRate, error)
}
