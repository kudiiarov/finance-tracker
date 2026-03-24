package domain

// Currency represents a currency with its precision.
type Currency struct {
	Code      string
	Name      string
	Precision int // 2 for fiat (cents), 8 for crypto (satoshis), 3 for metals (milligrams)
}

// currencyRegistry is the single source of truth for all currencies.
// To add a new currency, just add it here. Everything else updates automatically.
var currencyRegistry = map[string]Currency{
	// Fiat currencies
	"USD": {Code: "USD", Name: "US Dollar", Precision: 2},
	"EUR": {Code: "EUR", Name: "Euro", Precision: 2},
	"RUB": {Code: "RUB", Name: "Russian Ruble", Precision: 2},
	"CNY": {Code: "CNY", Name: "Chinese Yuan", Precision: 2},

	// Cryptocurrencies
	"BTC": {Code: "BTC", Name: "Bitcoin", Precision: 8},
	"ETH": {Code: "ETH", Name: "Ethereum", Precision: 8},

	// Precious metals (in grams, precision 3 for milligrams)
	"GOLD":   {Code: "GOLD", Name: "Gold (grams)", Precision: 3},
	"SILVER": {Code: "SILVER", Name: "Silver (grams)", Precision: 3},
}

// GetCurrency retrieves a currency by code.
// Returns (Currency, true) if found, (zero Currency, false) if not.
func GetCurrency(code string) (Currency, bool) {
	c, ok := currencyRegistry[code]
	return c, ok
}

// MustGetCurrency retrieves a currency by code.
// Panics if currency not found. Use only when you're sure it exists!
func MustGetCurrency(code string) Currency {
	c, ok := currencyRegistry[code]
	if !ok {
		panic("unknown currency: " + code)
	}
	return c
}

// IsValid checks if currency code exists in the registry.
func (c Currency) IsValid() bool {
	_, ok := currencyRegistry[c.Code]
	return ok
}

// SupportedCurrencies returns all available currencies from the registry.
func SupportedCurrencies() []Currency {
	currencies := make([]Currency, 0, len(currencyRegistry))
	for _, c := range currencyRegistry {
		currencies = append(currencies, c)
	}
	return currencies
}

// ToSmallestUnit converts human-readable amount to internal int64 format.
// Example: 0.000003 BTC → 300, 10.50 USD → 1050, 5.123 GOLD → 5123
func (c Currency) ToSmallestUnit(value float64) int64 {
	multiplier := 1.0
	for i := 0; i < c.Precision; i++ {
		multiplier *= 10
	}
	return int64(value * multiplier)
}

// FromSmallestUnit converts internal int64 to human-readable format.
// Example: 300 → 0.000003, 1050 → 10.50, 5123 → 5.123
func (c Currency) FromSmallestUnit(amount int64) float64 {
	multiplier := 1.0
	for i := 0; i < c.Precision; i++ {
		multiplier *= 10
	}
	return float64(amount) / multiplier
}
