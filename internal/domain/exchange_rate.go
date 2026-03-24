package domain

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// ExchangeRate represents a conversion rate between two currencies.
type ExchangeRate struct {
	ID           uuid.UUID `json:"id"`
	FromCurrency string    `json:"from_currency"`
	ToCurrency   string    `json:"to_currency"`
	Rate         float64   `json:"rate"`
	Source       string    `json:"source"`
	FetchedAt    int64     `json:"fetched_at"` // Unix timestamp
}

// IsValid checks if the rate is valid (positive and currencies are supported).
func (er ExchangeRate) IsValid() error {
	if er.Rate <= 0 {
		return errors.New("exchange rate must be positive")
	}
	fromCurrency := Currency{Code: er.FromCurrency}
	toCurrency := Currency{Code: er.ToCurrency}
	if !fromCurrency.IsValid() {
		return fmt.Errorf("invalid from_currency: %s", er.FromCurrency)
	}
	if !toCurrency.IsValid() {
		return fmt.Errorf("invalid to_currency: %s", er.ToCurrency)
	}
	return nil
}

// Convert calculates the converted amount.
// amount is in smallest units of FromCurrency.
// Returns amount in smallest units of ToCurrency.
func (er ExchangeRate) Convert(amount int64, fromPrecision, toPrecision int) int64 {
	// Convert from smallest units to human-readable
	fromMultiplier := 1.0
	for i := 0; i < fromPrecision; i++ {
		fromMultiplier *= 10
	}
	humanAmount := float64(amount) / fromMultiplier

	// Apply exchange rate
	convertedHuman := humanAmount * er.Rate

	// Convert to smallest units of target currency
	toMultiplier := 1.0
	for i := 0; i < toPrecision; i++ {
		toMultiplier *= 10
	}

	return int64(convertedHuman * toMultiplier)
}
