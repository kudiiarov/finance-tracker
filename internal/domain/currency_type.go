package domain

// CurrencyType categorizes currencies by their nature and provider requirements.
type CurrencyType string

const (
	CurrencyTypeFiat      CurrencyType = "fiat"      // USD, EUR, RUB, CNY
	CurrencyTypeCrypto    CurrencyType = "crypto"    // BTC, ETH
	CurrencyTypeMetal     CurrencyType = "metal"     // GOLD, SILVER
	CurrencyTypeCommodity CurrencyType = "commodity" // Reserved for future
	CurrencyTypeStock     CurrencyType = "stock"     // Reserved for future
)

// IsValid checks if the currency type is known.
func (ct CurrencyType) IsValid() bool {
	switch ct {
	case CurrencyTypeFiat, CurrencyTypeCrypto, CurrencyTypeMetal, CurrencyTypeCommodity, CurrencyTypeStock:
		return true
	}
	return false
}

// ProviderPreference defines which provider to use for a currency.
type ProviderPreference struct {
	Provider string `json:"provider" db:"provider"` // e.g., "coingecko"
	Priority int    `json:"priority" db:"priority"` // 1 = primary, 2 = fallback 1, etc.
}

// CurrencyInfo is the database model for configurable currencies.
type CurrencyInfo struct {
	Code              string               `json:"code" db:"code"`
	Name              string               `json:"name" db:"name"`
	Precision         int                  `json:"precision" db:"precision"`
	Type              CurrencyType         `json:"type" db:"currency_type"`
	Symbol            string               `json:"symbol" db:"symbol"`
	BaseUnit          string               `json:"base_unit" db:"base_unit"` // e.g., "gram" for metals
	IsActive          bool                 `json:"is_active" db:"is_active"`
	ProviderPrefs     []ProviderPreference `json:"provider_prefs" db:"-"` // Fetched separately
	ProviderPrefsJSON string               `json:"-" db:"provider_prefs"` // Stored as JSON
	CreatedAt         int64                `json:"created_at" db:"created_at"`
	UpdatedAt         int64                `json:"updated_at" db:"updated_at"`
}

// GetPrimaryProvider returns the primary provider (priority 1).
func (c *CurrencyInfo) GetPrimaryProvider() string {
	for _, p := range c.ProviderPrefs {
		if p.Priority == 1 {
			return p.Provider
		}
	}
	return ""
}

// GetFallbackProviders returns all fallback providers (priority > 1), ordered by priority.
func (c *CurrencyInfo) GetFallbackProviders() []string {
	var fallbacks []string
	for _, p := range c.ProviderPrefs {
		if p.Priority > 1 {
			fallbacks = append(fallbacks, p.Provider)
		}
	}
	return fallbacks
}

// ToCurrency converts CurrencyInfo to the simple Currency domain model.
func (c *CurrencyInfo) ToCurrency() Currency {
	return Currency{
		Code:      c.Code,
		Name:      c.Name,
		Precision: c.Precision,
	}
}
