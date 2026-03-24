package dto

// UpdateMainCurrencyRequest updates the main currency.
type UpdateMainCurrencyRequest struct {
	Currency string `json:"currency"`
}

// UpdateSettingsRequest updates all user settings.
type UpdateSettingsRequest struct {
	MainCurrency string `json:"main_currency"`
}
