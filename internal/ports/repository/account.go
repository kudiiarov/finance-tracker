// Package repository defines repository interfaces (ports in Clean Architecture).
// These interfaces are implemented by persistence adapters.
package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/sereozha/finance-tracker/internal/domain"
)

// AccountRepository defines storage operations for accounts.
type AccountRepository interface {
	Create(ctx context.Context, account *domain.Account) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Account, error)
	GetByUserID(ctx context.Context, userID string) ([]domain.Account, error)
	Update(ctx context.Context, account *domain.Account) error
	Delete(ctx context.Context, id uuid.UUID) error
}
