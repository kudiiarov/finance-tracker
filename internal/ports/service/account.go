// Package service defines service interfaces (driving ports in Clean Architecture).
// These interfaces are implemented by application services.
package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/sereozha/finance-tracker/internal/domain"
)

// AccountService defines business operations for accounts.
type AccountService interface {
	Create(ctx context.Context, userID string, input CreateAccountInput) (*domain.Account, error)
	Get(ctx context.Context, userID string, accountID uuid.UUID) (*domain.Account, error)
	List(ctx context.Context, userID string) ([]domain.Account, error)
	Update(ctx context.Context, userID string, accountID uuid.UUID, input UpdateAccountInput) (*domain.Account, error)
	Delete(ctx context.Context, userID string, accountID uuid.UUID) error
}

// CreateAccountInput contains data needed to create an account.
type CreateAccountInput struct {
	Name        string
	Type        domain.AccountType
	Currency    domain.Currency
	Amount      int64
	Description string
}

// UpdateAccountInput contains fields that can be updated.
// Pointers indicate optional fields - nil means "don't change".
type UpdateAccountInput struct {
	Name        *string
	Type        *domain.AccountType
	Amount      *int64
	Description *string
}
