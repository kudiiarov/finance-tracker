// Package dto contains request and response structures.
package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/sereozha/finance-tracker/internal/domain"
)

// CreateAccountRequest is what the client sends to create an account.
type CreateAccountRequest struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Currency    string `json:"currency"`
	Amount      int64  `json:"amount"`
	Description string `json:"description,omitempty"`
}

// AccountResponse is what we return to the client.
type AccountResponse struct {
	ID          uuid.UUID `json:"id"`
	UserID      string    `json:"user_id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Currency    string    `json:"currency"`
	Precision   int       `json:"precision"`
	Amount      int64     `json:"amount"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Conversion to user's main currency
	MainCurrency    string  `json:"main_currency"`
	ConvertedAmount int64   `json:"converted_amount"`
	ConvertedFloat  float64 `json:"converted_float"`
	ExchangeRate    float64 `json:"exchange_rate"`
}

// UpdateAccountRequest is what the client sends to update an account.
type UpdateAccountRequest struct {
	Name        *string `json:"name,omitempty"`
	Type        *string `json:"type,omitempty"`
	Amount      *int64  `json:"amount,omitempty"`
	Description *string `json:"description,omitempty"`
}

// ListAccountsResponse wraps the list.
type ListAccountsResponse struct {
	Accounts []AccountResponse `json:"accounts"`
}

// ToAccountResponse converts domain.Account to AccountResponse.
func ToAccountResponse(a domain.Account, mainCurrency domain.Currency, rate float64) AccountResponse {
	resp := AccountResponse{
		ID:           a.ID,
		UserID:       a.UserID,
		Name:         a.Name,
		Type:         string(a.Type),
		Currency:     a.Currency.Code,
		Precision:    a.Currency.Precision,
		Amount:       a.Amount,
		Description:  a.Description,
		CreatedAt:    a.CreatedAt,
		UpdatedAt:    a.UpdatedAt,
		MainCurrency: mainCurrency.Code,
	}

	if a.Currency.Code == mainCurrency.Code {
		resp.ConvertedAmount = a.Amount
		resp.ConvertedFloat = a.Currency.FromSmallestUnit(a.Amount)
		resp.ExchangeRate = 1.0
	} else {
		converted := domain.ExchangeRate{
			FromCurrency: a.Currency.Code,
			ToCurrency:   mainCurrency.Code,
			Rate:         rate,
		}.Convert(a.Amount, a.Currency.Precision, mainCurrency.Precision)

		resp.ConvertedAmount = converted
		resp.ConvertedFloat = mainCurrency.FromSmallestUnit(converted)
		resp.ExchangeRate = rate
	}

	return resp
}

// ToAccountResponseList converts a slice of accounts.
func ToAccountResponseList(accounts []domain.Account, mainCurrency domain.Currency, rates map[string]float64) ListAccountsResponse {
	response := make([]AccountResponse, len(accounts))
	for i, a := range accounts {
		rate, ok := rates[a.Currency.Code]
		if !ok {
			rate = 1.0
		}
		response[i] = ToAccountResponse(a, mainCurrency, rate)
	}
	return ListAccountsResponse{Accounts: response}
}
