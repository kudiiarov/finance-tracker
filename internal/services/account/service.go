// Package account provides account management business logic.
package account

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sereozha/finance-tracker/internal/domain"
	"github.com/sereozha/finance-tracker/internal/ports/repository"
	"github.com/sereozha/finance-tracker/internal/ports/service"
)

// Service implements service.AccountService.
type Service struct {
	repo repository.AccountRepository
}

// NewService creates a new account service.
func NewService(repo repository.AccountRepository) service.AccountService {
	return &Service{repo: repo}
}

// Create creates a new account with validation.
func (s *Service) Create(ctx context.Context, userID string, input service.CreateAccountInput) (*domain.Account, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("account name is required")
	}

	if !input.Currency.IsValid() {
		return nil, fmt.Errorf("invalid currency: %s", input.Currency.Code)
	}

	account := &domain.Account{
		ID:          uuid.New(),
		UserID:      userID,
		Name:        input.Name,
		Type:        input.Type,
		Currency:    input.Currency,
		Amount:      input.Amount,
		Description: input.Description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := account.Validate(); err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, account); err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	return account, nil
}

// Get retrieves an account by ID, checking ownership.
func (s *Service) Get(ctx context.Context, userID string, accountID uuid.UUID) (*domain.Account, error) {
	account, err := s.repo.GetByID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("account not found")
	}

	if account.UserID != userID {
		return nil, fmt.Errorf("account not found")
	}

	return account, nil
}

// List returns all accounts for a user.
func (s *Service) List(ctx context.Context, userID string) ([]domain.Account, error) {
	return s.repo.GetByUserID(ctx, userID)
}

// Update modifies an existing account.
func (s *Service) Update(ctx context.Context, userID string, accountID uuid.UUID, input service.UpdateAccountInput) (*domain.Account, error) {
	account, err := s.Get(ctx, userID, accountID)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		account.Name = *input.Name
	}
	if input.Type != nil {
		account.Type = *input.Type
	}
	if input.Amount != nil {
		account.Amount = *input.Amount
	}
	if input.Description != nil {
		account.Description = *input.Description
	}
	account.UpdatedAt = time.Now()

	if err := account.Validate(); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, account); err != nil {
		return nil, fmt.Errorf("failed to update account: %w", err)
	}

	return account, nil
}

// Delete removes an account after ownership check.
func (s *Service) Delete(ctx context.Context, userID string, accountID uuid.UUID) error {
	if _, err := s.Get(ctx, userID, accountID); err != nil {
		return err
	}
	return s.repo.Delete(ctx, accountID)
}
