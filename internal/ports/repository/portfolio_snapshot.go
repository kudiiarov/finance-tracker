package repository

import (
	"context"

	"github.com/sereozha/finance-tracker/internal/domain"
)

// PortfolioSnapshotRepository manages historical portfolio snapshots.
type PortfolioSnapshotRepository interface {
	// CreateSnapshot saves a new snapshot with all account details.
	CreateSnapshot(ctx context.Context, snapshot *domain.PortfolioSnapshot) error

	// GetSnapshot retrieves a specific day's snapshot for a user.
	GetSnapshot(ctx context.Context, userID string, date string) (*domain.PortfolioSnapshot, error)

	// GetLatestSnapshot gets the most recent snapshot for a user.
	GetLatestSnapshot(ctx context.Context, userID string) (*domain.PortfolioSnapshot, error)

	// GetHistory retrieves snapshots within a date range.
	GetHistory(ctx context.Context, query domain.PortfolioHistoryQuery) ([]domain.PortfolioSnapshot, error)

	// GetSnapshotAccounts retrieves account details for a snapshot.
	GetSnapshotAccounts(ctx context.Context, snapshotID string) ([]domain.PortfolioSnapshotAccount, error)

	// DeleteSnapshot removes a snapshot (for admin/reconciliation).
	DeleteSnapshot(ctx context.Context, snapshotID string) error

	// HasSnapshot checks if a snapshot exists for given date.
	HasSnapshot(ctx context.Context, userID string, date string) (bool, error)
}
