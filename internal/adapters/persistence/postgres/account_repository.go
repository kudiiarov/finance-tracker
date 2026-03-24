// Package postgres provides PostgreSQL repository implementations.
package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sereozha/finance-tracker/internal/domain"
	"github.com/sereozha/finance-tracker/internal/ports/repository"
)

// AccountRepository implements repository.AccountRepository using PostgreSQL.
type AccountRepository struct {
	pool *pgxpool.Pool
}

// NewAccountRepository creates a new repository instance.
func NewAccountRepository(pool *pgxpool.Pool) repository.AccountRepository {
	return &AccountRepository{pool: pool}
}

// Create inserts a new account.
func (r *AccountRepository) Create(ctx context.Context, account *domain.Account) error {
	query := `
		INSERT INTO accounts (id, user_id, name, type, currency, amount, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.pool.Exec(ctx, query,
		account.ID,
		account.UserID,
		account.Name,
		account.Type,
		account.Currency.Code,
		account.Amount,
		account.Description,
		account.CreatedAt,
		account.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create account: %w", err)
	}

	return nil
}

// GetByID retrieves an account by ID.
func (r *AccountRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Account, error) {
	query := `
		SELECT id, user_id, name, type, currency, amount, description, created_at, updated_at
		FROM accounts
		WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)
	return r.scanAccount(row)
}

// GetByUserID retrieves all accounts for a user.
func (r *AccountRepository) GetByUserID(ctx context.Context, userID string) ([]domain.Account, error) {
	query := `
		SELECT id, user_id, name, type, currency, amount, description, created_at, updated_at
		FROM accounts
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query accounts: %w", err)
	}
	defer rows.Close()

	var accounts []domain.Account
	for rows.Next() {
		account, err := r.scanAccount(rows)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, *account)
	}

	return accounts, rows.Err()
}

// Update modifies an existing account.
func (r *AccountRepository) Update(ctx context.Context, account *domain.Account) error {
	query := `
		UPDATE accounts
		SET name = $1, type = $2, currency = $3, amount = $4, description = $5, updated_at = $6
		WHERE id = $7
	`

	result, err := r.pool.Exec(ctx, query,
		account.Name,
		account.Type,
		account.Currency.Code,
		account.Amount,
		account.Description,
		account.UpdatedAt,
		account.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update account: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("account not found")
	}

	return nil
}

// Delete removes an account.
func (r *AccountRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM accounts WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete account: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("account not found")
	}

	return nil
}

// scanAccount scans a row into an Account struct.
func (r *AccountRepository) scanAccount(row pgx.Row) (*domain.Account, error) {
	var account domain.Account
	var currencyCode string

	err := row.Scan(
		&account.ID,
		&account.UserID,
		&account.Name,
		&account.Type,
		&currencyCode,
		&account.Amount,
		&account.Description,
		&account.CreatedAt,
		&account.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("account not found")
		}
		return nil, fmt.Errorf("failed to scan account: %w", err)
	}

	account.Currency = getCurrencyByCode(currencyCode)
	return &account, nil
}

func getCurrencyByCode(code string) domain.Currency {
	if c, ok := domain.GetCurrency(code); ok {
		return c
	}
	return domain.Currency{Code: code, Precision: 2}
}
