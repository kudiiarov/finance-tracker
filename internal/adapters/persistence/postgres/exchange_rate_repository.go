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

// ExchangeRateRepository implements repository.ExchangeRateRepository.
type ExchangeRateRepository struct {
	pool *pgxpool.Pool
}

// NewExchangeRateRepository creates a new repository instance.
func NewExchangeRateRepository(pool *pgxpool.Pool) repository.ExchangeRateRepository {
	return &ExchangeRateRepository{pool: pool}
}

// Save stores a new exchange rate.
func (r *ExchangeRateRepository) Save(ctx context.Context, rate *domain.ExchangeRate) error {
	if rate.ID == uuid.Nil {
		rate.ID = uuid.New()
	}

	query := `
		INSERT INTO exchange_rates (id, from_currency, to_currency, rate, source, fetched_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.pool.Exec(ctx, query,
		rate.ID,
		rate.FromCurrency,
		rate.ToCurrency,
		rate.Rate,
		rate.Source,
		rate.FetchedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save exchange rate: %w", err)
	}

	return nil
}

// GetLatest returns the most recent rate for a currency pair.
func (r *ExchangeRateRepository) GetLatest(ctx context.Context, fromCurrency, toCurrency string) (*domain.ExchangeRate, error) {
	query := `
		SELECT id, from_currency, to_currency, rate, source, fetched_at
		FROM exchange_rates
		WHERE from_currency = $1 AND to_currency = $2
		ORDER BY fetched_at DESC
		LIMIT 1
	`

	row := r.pool.QueryRow(ctx, query, fromCurrency, toCurrency)
	return r.scanRate(row)
}

// GetRateAt returns the rate closest to a specific time.
func (r *ExchangeRateRepository) GetRateAt(ctx context.Context, fromCurrency, toCurrency string, at int64) (*domain.ExchangeRate, error) {
	query := `
		SELECT id, from_currency, to_currency, rate, source, fetched_at
		FROM exchange_rates
		WHERE from_currency = $1 AND to_currency = $2
		ORDER BY ABS(fetched_at - $3)
		LIMIT 1
	`

	row := r.pool.QueryRow(ctx, query, fromCurrency, toCurrency, at)
	return r.scanRate(row)
}

// GetHistory returns rates within a time range.
func (r *ExchangeRateRepository) GetHistory(ctx context.Context, fromCurrency, toCurrency string, from, to int64) ([]domain.ExchangeRate, error) {
	query := `
		SELECT id, from_currency, to_currency, rate, source, fetched_at
		FROM exchange_rates
		WHERE from_currency = $1 AND to_currency = $2
		  AND fetched_at BETWEEN $3 AND $4
		ORDER BY fetched_at ASC
	`

	rows, err := r.pool.Query(ctx, query, fromCurrency, toCurrency, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to query history: %w", err)
	}
	defer rows.Close()

	var rates []domain.ExchangeRate
	for rows.Next() {
		rate, err := r.scanRate(rows)
		if err != nil {
			return nil, err
		}
		rates = append(rates, *rate)
	}

	return rates, rows.Err()
}

// DeleteOlderThan removes rates older than the specified time.
func (r *ExchangeRateRepository) DeleteOlderThan(ctx context.Context, before int64) error {
	query := `DELETE FROM exchange_rates WHERE fetched_at < $1`

	_, err := r.pool.Exec(ctx, query, before)
	if err != nil {
		return fmt.Errorf("failed to delete old rates: %w", err)
	}

	return nil
}

func (r *ExchangeRateRepository) scanRate(row pgx.Row) (*domain.ExchangeRate, error) {
	var rate domain.ExchangeRate

	err := row.Scan(
		&rate.ID,
		&rate.FromCurrency,
		&rate.ToCurrency,
		&rate.Rate,
		&rate.Source,
		&rate.FetchedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("exchange rate not found")
		}
		return nil, fmt.Errorf("failed to scan rate: %w", err)
	}

	return &rate, nil
}
