package dto

import (
	"github.com/sereozha/finance-tracker/internal/domain"
)

// CreateCurrencyRequest creates a new currency.
type CreateCurrencyRequest struct {
	Code          string                      `json:"code"`
	Name          string                      `json:"name"`
	Precision     int                         `json:"precision"`
	Type          domain.CurrencyType         `json:"type"`
	Symbol        string                      `json:"symbol"`
	BaseUnit      string                      `json:"base_unit"`
	ProviderPrefs []domain.ProviderPreference `json:"provider_prefs"`
}

// UpdateCurrencyRequest updates a currency.
type UpdateCurrencyRequest struct {
	Name      *string `json:"name,omitempty"`
	Precision *int    `json:"precision,omitempty"`
	Symbol    *string `json:"symbol,omitempty"`
	BaseUnit  *string `json:"base_unit,omitempty"`
	IsActive  *bool   `json:"is_active,omitempty"`
}

// UpdateProvidersRequest updates provider preferences.
type UpdateProvidersRequest struct {
	Providers []domain.ProviderPreference `json:"providers"`
}

// CurrencyInfoResponse is the API response for currency info.
type CurrencyInfoResponse struct {
	Code          string                      `json:"code"`
	Name          string                      `json:"name"`
	Precision     int                         `json:"precision"`
	Type          domain.CurrencyType         `json:"type"`
	Symbol        string                      `json:"symbol"`
	BaseUnit      string                      `json:"base_unit"`
	IsActive      bool                        `json:"is_active"`
	ProviderPrefs []domain.ProviderPreference `json:"provider_prefs"`
	CreatedAt     int64                       `json:"created_at"`
	UpdatedAt     int64                       `json:"updated_at"`

	// Computed
	PrimaryProvider   string   `json:"primary_provider"`
	FallbackProviders []string `json:"fallback_providers"`
}

// ToCurrencyInfoResponse converts domain to DTO.
func ToCurrencyInfoResponse(c *domain.CurrencyInfo) CurrencyInfoResponse {
	return CurrencyInfoResponse{
		Code:              c.Code,
		Name:              c.Name,
		Precision:         c.Precision,
		Type:              c.Type,
		Symbol:            c.Symbol,
		BaseUnit:          c.BaseUnit,
		IsActive:          c.IsActive,
		ProviderPrefs:     c.ProviderPrefs,
		CreatedAt:         c.CreatedAt,
		UpdatedAt:         c.UpdatedAt,
		PrimaryProvider:   c.GetPrimaryProvider(),
		FallbackProviders: c.GetFallbackProviders(),
	}
}

// ToCurrencyInfoList converts a list.
func ToCurrencyInfoList(currencies []domain.CurrencyInfo) []CurrencyInfoResponse {
	result := make([]CurrencyInfoResponse, len(currencies))
	for i, c := range currencies {
		result[i] = ToCurrencyInfoResponse(&c)
	}
	return result
}
