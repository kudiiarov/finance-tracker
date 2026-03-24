package domain

import "time"

// ProviderCapability describes what a provider can fetch.
type ProviderCapability struct {
	Name           string         // Provider identifier: "coingecko"
	Types          []CurrencyType // Currency types supported: [crypto]
	BaseCurrencies []string       // Can provide rates TO these: [USD, USDT]
	DirectPairs    []CurrencyPair // Specific supported pairs
	RateLimit      RateLimitConfig
	Priority       int // Default priority (lower = preferred)
}

// CurrencyPair represents a specific from/to pair.
type CurrencyPair struct {
	From string
	To   string
}

// RateLimitConfig defines rate limiting for a provider.
type RateLimitConfig struct {
	RequestsPerMinute int
	RequestsPerHour   int
	RequestsPerDay    int
}

// ProviderStatus tracks provider health.
type ProviderStatus struct {
	Name          string
	IsHealthy     bool
	LastSuccess   time.Time
	LastFailure   time.Time
	FailureCount  int
	SuccessCount  int
	AvgResponseMs int64
}

// RateRequest is a request to fetch rates.
type RateRequest struct {
	FromCurrency       string
	ToCurrency         string
	Currencies         []string // Currencies to fetch rates for (optional)
	PreferredProviders []string // Optional: force specific providers
}

// ProviderHealthChecker monitors provider health.
type ProviderHealthChecker interface {
	RecordSuccess(provider string, duration time.Duration)
	RecordFailure(provider string, err error)
	IsHealthy(provider string) bool
	GetStatus(provider string) ProviderStatus
}
