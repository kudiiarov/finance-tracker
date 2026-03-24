package exchange

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/sereozha/finance-tracker/internal/domain"
	"github.com/sereozha/finance-tracker/internal/ports/logger"
)

// nopLogger is a no-op logger used when no logger is set.
type nopLogger struct{}

func (n nopLogger) Debug(msg string, fields ...logger.Field)      {}
func (n nopLogger) Info(msg string, fields ...logger.Field)       {}
func (n nopLogger) Warn(msg string, fields ...logger.Field)       {}
func (n nopLogger) Error(msg string, fields ...logger.Field)      {}
func (n nopLogger) Fatal(msg string, fields ...logger.Field)      {}
func (n nopLogger) With(fields ...logger.Field) logger.Logger     { return n }
func (n nopLogger) WithContext(ctx context.Context) logger.Logger { return n }
func (n nopLogger) Sync() error                                   { return nil }

// MetalsAPIProvider fetches precious metals prices from GoldAPI.io.
type MetalsAPIProvider struct {
	client  *http.Client
	baseURL string
	apiKey  string
	cap     domain.ProviderCapability
	log     logger.Logger
}

// NewMetalsAPIProvider creates a new precious metals provider.
func NewMetalsAPIProvider() *MetalsAPIProvider {
	apiKey := os.Getenv("GOLDAPI_KEY")

	return &MetalsAPIProvider{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://www.goldapi.io/api",
		apiKey:  apiKey,
		cap: domain.ProviderCapability{
			Name:           "goldapi.io",
			Types:          []domain.CurrencyType{domain.CurrencyTypeMetal},
			BaseCurrencies: []string{"USD"},
			Priority:       1,
			RateLimit: domain.RateLimitConfig{
				RequestsPerMinute: 20,
				RequestsPerDay:    100,
			},
		},
	}
}

// SetLogger sets the logger for this provider (called during initialization).
func (p *MetalsAPIProvider) SetLogger(log logger.Logger) {
	p.log = log.With(logger.String(logger.FieldComponent, "goldapi_provider"))
}

// getLogger returns the configured logger or a no-op logger if none is set.
func (p *MetalsAPIProvider) getLogger() logger.Logger {
	if p.log == nil {
		return nopLogger{}
	}
	return p.log
}

// GetCapabilities returns what this provider can fetch.
func (p *MetalsAPIProvider) GetCapabilities() domain.ProviderCapability {
	return p.cap
}

// metalMapping maps currency codes to GoldAPI metal codes.
var metalMapping = map[string]string{
	"GOLD":   "XAU",
	"SILVER": "XAG",
	"PLATIN": "XPT",
	"PALLAD": "XPD",
}

// FetchRates fetches precious metals rates from GoldAPI.io.
func (p *MetalsAPIProvider) FetchRates(ctx context.Context, request domain.RateRequest) ([]domain.ExchangeRate, error) {
	if p.apiKey == "" {
		return nil, fmt.Errorf("GOLDAPI_KEY not set")
	}

	now := time.Now().Unix()
	rates := make([]domain.ExchangeRate, 0)
	gramsPerTroyOz := 31.1034768
	var lastErr error

	// Determine which metals to fetch
	metalsToFetch := []string{"GOLD", "SILVER"}
	if len(request.Currencies) > 0 {
		metalsToFetch = request.Currencies
	}

	for _, currency := range metalsToFetch {
		metalCode, ok := metalMapping[currency]
		if !ok {
			p.getLogger().Warn("unknown_metal_currency", logger.String(logger.FieldCurrency, currency))
			continue
		}

		metalRate, err := p.fetchMetalPrice(ctx, metalCode, "USD")
		if err != nil {
			p.getLogger().Warn("failed_to_fetch_metal_price",
				logger.String(logger.FieldCurrency, currency),
				logger.String("metal_code", metalCode),
				logger.Error(err))
			lastErr = err
			continue
		}

		ratePerGram := metalRate / gramsPerTroyOz
		rates = append(rates, domain.ExchangeRate{
			ID:           uuid.New(),
			FromCurrency: currency,
			ToCurrency:   "USD",
			Rate:         ratePerGram,
			Source:       p.cap.Name,
			FetchedAt:    now,
		})

		p.getLogger().Debug("metal_price_fetched",
			logger.String(logger.FieldCurrency, currency),
			logger.Float64("price_per_oz", metalRate),
			logger.Float64("price_per_gram", ratePerGram))
	}

	// If we got no rates and had errors, return the last error
	if len(rates) == 0 && lastErr != nil {
		return nil, fmt.Errorf("failed to fetch any metal rates: %w", lastErr)
	}

	// Log summary if some but not all succeeded
	if len(rates) > 0 && len(rates) < len(metalsToFetch) {
		p.getLogger().Warn("partial_metal_fetch_success",
			logger.Int("requested", len(metalsToFetch)),
			logger.Int("fetched", len(rates)))
	}

	return rates, nil
}

type goldAPIResponse struct {
	Timestamp int64   `json:"timestamp"`
	Metal     string  `json:"metal"`
	Currency  string  `json:"currency"`
	Exchange  string  `json:"exchange"`
	Symbol    string  `json:"symbol"`
	PrevClose float64 `json:"prev_close_price"`
	Open      float64 `json:"open_price"`
	Low       float64 `json:"low_price"`
	High      float64 `json:"high_price"`
	Price     float64 `json:"price"`
	Ch        float64 `json:"ch"`
	Chp       float64 `json:"chp"`
	Ask       float64 `json:"ask"`
	Bid       float64 `json:"bid"`
}

// goldAPIError represents an error response from GoldAPI.
type goldAPIError struct {
	Error string `json:"error"`
}

// extractErrorMessage parses the GoldAPI error response and extracts the message.
func extractErrorMessage(body []byte) string {
	var apiErr goldAPIError
	if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.Error != "" {
		return apiErr.Error
	}
	// Fallback to raw body if not valid JSON
	return string(body)
}

func (p *MetalsAPIProvider) fetchMetalPrice(ctx context.Context, metal, currency string) (float64, error) {
	url := fmt.Sprintf("%s/%s/%s", p.baseURL, metal, currency)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("x-access-token", p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		errMsg := extractErrorMessage(body)

		p.getLogger().Warn("goldapi_error_response",
			logger.String("metal", metal),
			logger.Int("status_code", resp.StatusCode),
			logger.String("error_message", errMsg),
			logger.String("raw_response", string(body)))

		switch resp.StatusCode {
		case http.StatusForbidden:
			return 0, fmt.Errorf("API access forbidden (403): %s", errMsg)
		case http.StatusTooManyRequests:
			return 0, fmt.Errorf("rate limit exceeded (429): %s", errMsg)
		case http.StatusUnauthorized:
			return 0, fmt.Errorf("API key invalid (401): %s", errMsg)
		default:
			return 0, fmt.Errorf("API returned status %d: %s", resp.StatusCode, errMsg)
		}
	}

	var result goldAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Price <= 0 {
		return 0, fmt.Errorf("invalid price received: %f", result.Price)
	}

	return result.Price, nil
}
