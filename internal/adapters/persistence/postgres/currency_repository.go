package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sereozha/finance-tracker/internal/domain"
	"github.com/sereozha/finance-tracker/internal/ports/repository"
)

// CurrencyRepository implements repository.CurrencyRepository.
type CurrencyRepository struct {
	pool *pgxpool.Pool
}

// NewCurrencyRepository creates a new currency repository.
func NewCurrencyRepository(pool *pgxpool.Pool) repository.CurrencyRepository {
	return &CurrencyRepository{pool: pool}
}

// Create adds a new currency to the registry.
func (r *CurrencyRepository) Create(ctx context.Context, currency *domain.CurrencyInfo) error {
	prefsJSON, err := json.Marshal(currency.ProviderPrefs)
	if err != nil {
		return fmt.Errorf("failed to marshal provider prefs: %w", err)
	}

	now := time.Now().Unix()
	query := `
		INSERT INTO currencies (code, name, precision, currency_type, symbol, base_unit, is_active, provider_prefs, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err = r.pool.Exec(ctx, query,
		currency.Code,
		currency.Name,
		currency.Precision,
		currency.Type,
		currency.Symbol,
		currency.BaseUnit,
		currency.IsActive,
		prefsJSON,
		now,
		now,
	)
	if err != nil {
		return fmt.Errorf("failed to create currency: %w", err)
	}

	return nil
}

// Get retrieves a currency by code.
func (r *CurrencyRepository) Get(ctx context.Context, code string) (*domain.CurrencyInfo, error) {
	query := `
		SELECT code, name, precision, currency_type, symbol, base_unit, is_active, provider_prefs, created_at, updated_at
		FROM currencies
		WHERE code = $1
	`

	var c domain.CurrencyInfo
	var prefsJSON []byte

	row := r.pool.QueryRow(ctx, query, code)
	err := row.Scan(
		&c.Code,
		&c.Name,
		&c.Precision,
		&c.Type,
		&c.Symbol,
		&c.BaseUnit,
		&c.IsActive,
		&prefsJSON,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("currency not found: %s", code)
		}
		return nil, fmt.Errorf("failed to get currency: %w", err)
	}

	// Unmarshal provider prefs
	if len(prefsJSON) > 0 {
		if err := json.Unmarshal(prefsJSON, &c.ProviderPrefs); err != nil {
			return nil, fmt.Errorf("failed to unmarshal provider prefs: %w", err)
		}
	}

	return &c, nil
}

// Update updates currency details.
func (r *CurrencyRepository) Update(ctx context.Context, currency *domain.CurrencyInfo) error {
	query := `
		UPDATE currencies
		SET name = $1, precision = $2, symbol = $3, base_unit = $4, is_active = $5, updated_at = $6
		WHERE code = $7
	`

	result, err := r.pool.Exec(ctx, query,
		currency.Name,
		currency.Precision,
		currency.Symbol,
		currency.BaseUnit,
		currency.IsActive,
		time.Now().Unix(),
		currency.Code,
	)
	if err != nil {
		return fmt.Errorf("failed to update currency: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("currency not found: %s", currency.Code)
	}

	return nil
}

// UpdateProviderPrefs updates the provider preferences for a currency.
func (r *CurrencyRepository) UpdateProviderPrefs(ctx context.Context, code string, prefs []domain.ProviderPreference) error {
	prefsJSON, err := json.Marshal(prefs)
	if err != nil {
		return fmt.Errorf("failed to marshal provider prefs: %w", err)
	}

	query := `
		UPDATE currencies
		SET provider_prefs = $1, updated_at = $2
		WHERE code = $3
	`

	result, err := r.pool.Exec(ctx, query, prefsJSON, time.Now().Unix(), code)
	if err != nil {
		return fmt.Errorf("failed to update provider prefs: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("currency not found: %s", code)
	}

	return nil
}

// ListActive returns all active currencies.
func (r *CurrencyRepository) ListActive(ctx context.Context) ([]domain.CurrencyInfo, error) {
	query := `
		SELECT code, name, precision, currency_type, symbol, base_unit, is_active, provider_prefs, created_at, updated_at
		FROM currencies
		WHERE is_active = true
		ORDER BY code
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list currencies: %w", err)
	}
	defer rows.Close()

	return r.scanCurrencies(rows)
}

// ListByType returns currencies filtered by type.
func (r *CurrencyRepository) ListByType(ctx context.Context, currencyType domain.CurrencyType) ([]domain.CurrencyInfo, error) {
	query := `
		SELECT code, name, precision, currency_type, symbol, base_unit, is_active, provider_prefs, created_at, updated_at
		FROM currencies
		WHERE currency_type = $1 AND is_active = true
		ORDER BY code
	`

	rows, err := r.pool.Query(ctx, query, currencyType)
	if err != nil {
		return nil, fmt.Errorf("failed to list currencies by type: %w", err)
	}
	defer rows.Close()

	return r.scanCurrencies(rows)
}

// Delete marks a currency as inactive (soft delete).
func (r *CurrencyRepository) Delete(ctx context.Context, code string) error {
	query := `
		UPDATE currencies
		SET is_active = false, updated_at = $1
		WHERE code = $2
	`

	result, err := r.pool.Exec(ctx, query, time.Now().Unix(), code)
	if err != nil {
		return fmt.Errorf("failed to delete currency: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("currency not found: %s", code)
	}

	return nil
}

func (r *CurrencyRepository) scanCurrencies(rows pgx.Rows) ([]domain.CurrencyInfo, error) {
	var currencies []domain.CurrencyInfo

	for rows.Next() {
		var c domain.CurrencyInfo
		var prefsJSON []byte

		err := rows.Scan(
			&c.Code,
			&c.Name,
			&c.Precision,
			&c.Type,
			&c.Symbol,
			&c.BaseUnit,
			&c.IsActive,
			&prefsJSON,
			&c.CreatedAt,
			&c.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan currency: %w", err)
		}

		// Unmarshal provider prefs
		if len(prefsJSON) > 0 {
			if err := json.Unmarshal(prefsJSON, &c.ProviderPrefs); err != nil {
				return nil, fmt.Errorf("failed to unmarshal provider prefs: %w", err)
			}
		}

		currencies = append(currencies, c)
	}

	return currencies, rows.Err()
}

// Ensure sql import is used
var _ = sql.NullString{}
