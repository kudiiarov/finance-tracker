package dto

// CurrencyResponse represents a currency in the API.
type CurrencyResponse struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	Precision int    `json:"precision"`
}

// ListCurrenciesResponse wraps the currency list.
type ListCurrenciesResponse struct {
	Currencies []CurrencyResponse `json:"currencies"`
}
