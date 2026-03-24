package service

import (
	"context"
)

// StatisticsService defines financial statistics operations.
type StatisticsService interface {
	// GetTotalBalance calculates total balance across all accounts in target currency.
	GetTotalBalance(ctx context.Context, userID string, targetCurrency string) (*StatisticsResult, error)
}

// StatisticsResult represents total wealth calculation.
type StatisticsResult struct {
	UserID         string     `json:"user_id"`
	TargetCurrency string     `json:"target_currency"`
	TotalAmount    int64      `json:"total_amount"` // In smallest units
	TotalFloat     float64    `json:"total_float"`  // Human-readable
	Accounts       int        `json:"accounts"`
	RatesUsed      []RateInfo `json:"rates_used"`
}

// RateInfo describes which exchange rate was used for calculation.
type RateInfo struct {
	FromCurrency string  `json:"from_currency"`
	ToCurrency   string  `json:"to_currency"`
	Rate         float64 `json:"rate"`
	Source       string  `json:"source"`
}
