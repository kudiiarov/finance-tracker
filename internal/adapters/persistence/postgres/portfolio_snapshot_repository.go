package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sereozha/finance-tracker/internal/domain"
	"github.com/sereozha/finance-tracker/internal/ports/repository"
)

// PortfolioSnapshotRepository implements repository.PortfolioSnapshotRepository.
type PortfolioSnapshotRepository struct {
	pool *pgxpool.Pool
}

// NewPortfolioSnapshotRepository creates a new repository.
func NewPortfolioSnapshotRepository(pool *pgxpool.Pool) repository.PortfolioSnapshotRepository {
	return &PortfolioSnapshotRepository{pool: pool}
}

// CreateSnapshot saves a new snapshot with all account details.
func (r *PortfolioSnapshotRepository) CreateSnapshot(ctx context.Context, snapshot *domain.PortfolioSnapshot) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Insert main snapshot
	snapshotQuery := `
		INSERT INTO portfolio_snapshots 
		(user_id, snapshot_date, snapshot_timestamp, total_balance_usd, total_balance_float,
		 account_count, currency_count, calculation_duration_ms, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`

	now := time.Now().Unix()
	var snapshotID string
	err = tx.QueryRow(ctx, snapshotQuery,
		snapshot.UserID,
		snapshot.SnapshotDate,
		snapshot.SnapshotTimestamp,
		snapshot.TotalBalanceUSD,
		snapshot.TotalBalanceFloat,
		snapshot.AccountCount,
		snapshot.CurrencyCount,
		snapshot.CalculationDurationMs,
		now,
	).Scan(&snapshotID)
	if err != nil {
		return fmt.Errorf("failed to insert snapshot: %w", err)
	}

	// Insert account details
	accountQuery := `
		INSERT INTO portfolio_snapshot_accounts
		(snapshot_id, account_id, account_name, account_type, currency_code, currency_precision,
		 amount, amount_float, rate_to_usd, rate_source, rate_timestamp, 
		 converted_amount_usd, converted_amount_float, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	for _, acc := range snapshot.Accounts {
		_, err = tx.Exec(ctx, accountQuery,
			snapshotID,
			acc.AccountID,
			acc.AccountName,
			acc.AccountType,
			acc.CurrencyCode,
			acc.CurrencyPrecision,
			acc.Amount,
			acc.AmountFloat,
			acc.RateToUSD,
			acc.RateSource,
			acc.RateTimestamp,
			acc.ConvertedAmountUSD,
			acc.ConvertedAmountFloat,
			now,
		)
		if err != nil {
			return fmt.Errorf("failed to insert account snapshot: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	snapshot.ID = snapshotID
	return nil
}

// GetSnapshot retrieves a specific day's snapshot.
func (r *PortfolioSnapshotRepository) GetSnapshot(ctx context.Context, userID string, date string) (*domain.PortfolioSnapshot, error) {
	query := `
		SELECT id, user_id, snapshot_date, snapshot_timestamp, total_balance_usd, total_balance_float,
		       account_count, currency_count, calculation_duration_ms, created_at
		FROM portfolio_snapshots
		WHERE user_id = $1 AND snapshot_date = $2
	`

	row := r.pool.QueryRow(ctx, query, userID, date)
	return r.scanSnapshot(row)
}

// GetLatestSnapshot gets the most recent snapshot.
func (r *PortfolioSnapshotRepository) GetLatestSnapshot(ctx context.Context, userID string) (*domain.PortfolioSnapshot, error) {
	query := `
		SELECT id, user_id, snapshot_date, snapshot_timestamp, total_balance_usd, total_balance_float,
		       account_count, currency_count, calculation_duration_ms, created_at
		FROM portfolio_snapshots
		WHERE user_id = $1
		ORDER BY snapshot_date DESC
		LIMIT 1
	`

	row := r.pool.QueryRow(ctx, query, userID)
	return r.scanSnapshot(row)
}

// GetHistory retrieves snapshots within a date range.
func (r *PortfolioSnapshotRepository) GetHistory(ctx context.Context, query domain.PortfolioHistoryQuery) ([]domain.PortfolioSnapshot, error) {
	order := "ASC"
	if query.OrderDesc {
		order = "DESC"
	}

	sql := fmt.Sprintf(`
		SELECT id, user_id, snapshot_date, snapshot_timestamp, total_balance_usd, total_balance_float,
		       account_count, currency_count, calculation_duration_ms, created_at
		FROM portfolio_snapshots
		WHERE user_id = $1 AND snapshot_date >= $2 AND snapshot_date <= $3
		ORDER BY snapshot_date %s
	`, order)

	rows, err := r.pool.Query(ctx, sql, query.UserID, query.FromDate, query.ToDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query history: %w", err)
	}
	defer rows.Close()

	var snapshots []domain.PortfolioSnapshot
	for rows.Next() {
		snapshot, err := r.scanSnapshot(rows)
		if err != nil {
			return nil, err
		}
		snapshots = append(snapshots, *snapshot)
	}

	return snapshots, rows.Err()
}

// GetSnapshotAccounts retrieves account details for a snapshot.
func (r *PortfolioSnapshotRepository) GetSnapshotAccounts(ctx context.Context, snapshotID string) ([]domain.PortfolioSnapshotAccount, error) {
	query := `
		SELECT id, snapshot_id, account_id, account_name, account_type, currency_code, currency_precision,
		       amount, amount_float, rate_to_usd, rate_source, rate_timestamp,
		       converted_amount_usd, converted_amount_float, created_at
		FROM portfolio_snapshot_accounts
		WHERE snapshot_id = $1
		ORDER BY converted_amount_usd DESC
	`

	rows, err := r.pool.Query(ctx, query, snapshotID)
	if err != nil {
		return nil, fmt.Errorf("failed to query accounts: %w", err)
	}
	defer rows.Close()

	var accounts []domain.PortfolioSnapshotAccount
	for rows.Next() {
		acc, err := r.scanAccount(rows)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, *acc)
	}

	return accounts, rows.Err()
}

// DeleteSnapshot removes a snapshot.
func (r *PortfolioSnapshotRepository) DeleteSnapshot(ctx context.Context, snapshotID string) error {
	query := `DELETE FROM portfolio_snapshots WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, snapshotID)
	return err
}

// HasSnapshot checks if a snapshot exists.
func (r *PortfolioSnapshotRepository) HasSnapshot(ctx context.Context, userID string, date string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM portfolio_snapshots WHERE user_id = $1 AND snapshot_date = $2)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, userID, date).Scan(&exists)
	return exists, err
}

func (r *PortfolioSnapshotRepository) scanSnapshot(row pgx.Row) (*domain.PortfolioSnapshot, error) {
	var s domain.PortfolioSnapshot
	var snapshotDate time.Time

	err := row.Scan(
		&s.ID,
		&s.UserID,
		&snapshotDate,
		&s.SnapshotTimestamp,
		&s.TotalBalanceUSD,
		&s.TotalBalanceFloat,
		&s.AccountCount,
		&s.CurrencyCount,
		&s.CalculationDurationMs,
		&s.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("snapshot not found")
		}
		return nil, fmt.Errorf("failed to scan snapshot: %w", err)
	}

	s.SnapshotDate = snapshotDate.Format("2006-01-02")
	return &s, nil
}

func (r *PortfolioSnapshotRepository) scanAccount(row pgx.Row) (*domain.PortfolioSnapshotAccount, error) {
	var a domain.PortfolioSnapshotAccount
	err := row.Scan(
		&a.ID,
		&a.SnapshotID,
		&a.AccountID,
		&a.AccountName,
		&a.AccountType,
		&a.CurrencyCode,
		&a.CurrencyPrecision,
		&a.Amount,
		&a.AmountFloat,
		&a.RateToUSD,
		&a.RateSource,
		&a.RateTimestamp,
		&a.ConvertedAmountUSD,
		&a.ConvertedAmountFloat,
		&a.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan account: %w", err)
	}
	return &a, nil
}
