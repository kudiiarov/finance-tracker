package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// AccountType represents the type of financial account.
type AccountType string

const (
	AccountTypeCash       AccountType = "cash"
	AccountTypeBank       AccountType = "bank"
	AccountTypeCreditCard AccountType = "credit_card"
	AccountTypeCrypto     AccountType = "crypto"
	AccountTypeInvestment AccountType = "investment"
)

// Account represents a financial account.
type Account struct {
	ID          uuid.UUID   `json:"id"`
	UserID      string      `json:"user_id"`
	Name        string      `json:"name"`
	Type        AccountType `json:"type"`
	Currency    Currency    `json:"currency"`
	Amount      int64       `json:"amount"` // In smallest units (cents/satoshis)
	Description string      `json:"description,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// Validate checks business rules.
func (a *Account) Validate() error {
	if a.Name == "" {
		return errors.New("account name is required")
	}
	if !a.Currency.IsValid() {
		return errors.New("invalid currency")
	}
	return nil
}
