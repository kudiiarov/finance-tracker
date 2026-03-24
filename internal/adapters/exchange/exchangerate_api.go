package exchange

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/sereozha/finance-tracker/internal/domain"
)

// ExchangeRateAPIProvider fetches fiat exchange rates from exchangerate-api.com.
type ExchangeRateAPIProvider struct {
	client  *http.Client
	baseURL string
	cap     domain.ProviderCapability
}

// NewExchangeRateAPIProvider creates a new provider.
func NewExchangeRateAPIProvider() *ExchangeRateAPIProvider {
	return &ExchangeRateAPIProvider{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://api.exchangerate-api.com/v4/latest",
		cap: domain.ProviderCapability{
			Name:           "exchangerate-api",
			Types:          []domain.CurrencyType{domain.CurrencyTypeFiat},
			BaseCurrencies: []string{"USD"},
			Priority:       1,
			RateLimit: domain.RateLimitConfig{
				RequestsPerMinute: 60,
			},
		},
	}
}

// GetCapabilities returns what this provider can fetch.
func (p *ExchangeRateAPIProvider) GetCapabilities() domain.ProviderCapability {
	return p.cap
}

// FetchRates fetches fiat rates with USD as base.
func (p *ExchangeRateAPIProvider) FetchRates(ctx context.Context, request domain.RateRequest) ([]domain.ExchangeRate, error) {
	url := fmt.Sprintf("%s/USD", p.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var result struct {
		Base  string             `json:"base"`
		Date  string             `json:"date"`
		Rates map[string]float64 `json:"rates"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	now := time.Now().Unix()
	rates := make([]domain.ExchangeRate, 0)

	// Add USD to USD
	rates = append(rates, domain.ExchangeRate{
		ID:           uuid.New(),
		FromCurrency: "USD",
		ToCurrency:   "USD",
		Rate:         1.0,
		Source:       p.cap.Name,
		FetchedAt:    now,
	})

	// Convert rates: "1 USD = X EUR" → "1 EUR = 1/X USD"
	// Use requested currencies or default set
	currenciesWeNeed := []string{"EUR", "GBP", "RUB", "CNY"}
	if len(request.Currencies) > 0 {
		currenciesWeNeed = request.Currencies
	}

	for _, currency := range currenciesWeNeed {
		if usdToCurrency, ok := result.Rates[currency]; ok && usdToCurrency > 0 {
			toUSDRate := 1.0 / usdToCurrency

			rates = append(rates, domain.ExchangeRate{
				ID:           uuid.New(),
				FromCurrency: currency,
				ToCurrency:   "USD",
				Rate:         toUSDRate,
				Source:       p.cap.Name,
				FetchedAt:    now,
			})
		}
	}

	return rates, nil
}
