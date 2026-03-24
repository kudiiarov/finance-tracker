// Package statistics provides portfolio statistics and snapshot services.
package statistics

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/sereozha/finance-tracker/internal/domain"
	"github.com/sereozha/finance-tracker/internal/ports/logger"
	"github.com/sereozha/finance-tracker/internal/ports/repository"
)

// SnapshotService handles creation and retrieval of portfolio snapshots.
type SnapshotService struct {
	accountRepo  repository.AccountRepository
	snapshotRepo repository.PortfolioSnapshotRepository
	rateRepo     repository.ExchangeRateRepository
	currencyRepo repository.CurrencyRepository
	log          logger.Logger
}

// NewSnapshotService creates a new snapshot service.
func NewSnapshotService(
	accountRepo repository.AccountRepository,
	snapshotRepo repository.PortfolioSnapshotRepository,
	rateRepo repository.ExchangeRateRepository,
	currencyRepo repository.CurrencyRepository,
	log logger.Logger,
) *SnapshotService {
	return &SnapshotService{
		accountRepo:  accountRepo,
		snapshotRepo: snapshotRepo,
		rateRepo:     rateRepo,
		currencyRepo: currencyRepo,
		log:          log.With(logger.String(logger.FieldComponent, "snapshot_service")),
	}
}

// CreateDailySnapshot creates a snapshot for today (or specified date).
func (s *SnapshotService) CreateDailySnapshot(ctx context.Context, userID string, date string) (*domain.PortfolioSnapshot, error) {
	startTime := time.Now()

	// Check if snapshot already exists
	exists, err := s.snapshotRepo.HasSnapshot(ctx, userID, date)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing snapshot: %w", err)
	}
	if exists {
		s.log.Info("snapshot_already_exists",
			logger.String(logger.FieldUserID, userID),
			logger.String("date", date),
		)
		return s.snapshotRepo.GetSnapshot(ctx, userID, date)
	}

	s.log.Info("creating_daily_snapshot",
		logger.String(logger.FieldUserID, userID),
		logger.String("date", date),
	)

	// Get user's accounts
	accounts, err := s.accountRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %w", err)
	}

	if len(accounts) == 0 {
		s.log.Warn("no_accounts_for_snapshot",
			logger.String(logger.FieldUserID, userID),
		)
		return nil, fmt.Errorf("no accounts found for user")
	}

	// Parse date to timestamp (00:00:00 UTC)
	snapshotTime, _ := time.Parse("2006-01-02", date)
	snapshotTimestamp := snapshotTime.Unix()

	// Build snapshot
	snapshot := &domain.PortfolioSnapshot{
		UserID:            userID,
		SnapshotDate:      date,
		SnapshotTimestamp: snapshotTimestamp,
		AccountCount:      len(accounts),
		CreatedAt:         time.Now().Unix(),
	}

	// Track unique currencies
	currencyCodes := make(map[string]bool)
	var totalUSD int64

	// Process each account
	for _, account := range accounts {
		currencyCodes[account.Currency.Code] = true

		// Get rate for this currency to USD at snapshot time
		rate, rateSource, rateTimestamp, err := s.getRateAtTime(ctx, account.Currency.Code, "USD", snapshotTimestamp)
		if err != nil {
			s.log.Error("failed_to_get_rate",
				logger.Error(err),
				logger.String(logger.FieldCurrency, account.Currency.Code),
			)
			// Continue with 0 rate - account won't contribute to total
			rate = 0
		}

		// Convert amount to USD
		var convertedUSD int64
		if rate > 0 {
			// account.Amount is in currency's smallest units
			// rate is USD per 1 unit of currency
			// For crypto: amount in satoshis (1e-8), rate in USD per BTC
			// converted = amount * rate / 10^precision
			convertedFloat := float64(account.Amount) * rate
			for i := 0; i < account.Currency.Precision; i++ {
				convertedFloat /= 10
			}
			// Now convert to USD cents (precision 2)
			convertedUSD = int64(convertedFloat * 100)
		}

		accountSnapshot := domain.PortfolioSnapshotAccount{
			AccountID:            account.ID.String(),
			AccountName:          account.Name,
			AccountType:          string(account.Type),
			CurrencyCode:         account.Currency.Code,
			CurrencyPrecision:    account.Currency.Precision,
			Amount:               account.Amount,
			AmountFloat:          account.Currency.FromSmallestUnit(account.Amount),
			RateToUSD:            rate,
			RateSource:           rateSource,
			RateTimestamp:        rateTimestamp,
			ConvertedAmountUSD:   convertedUSD,
			ConvertedAmountFloat: float64(convertedUSD) / 100,
			CreatedAt:            time.Now().Unix(),
		}

		snapshot.Accounts = append(snapshot.Accounts, accountSnapshot)
		totalUSD += convertedUSD
	}

	snapshot.TotalBalanceUSD = totalUSD
	snapshot.TotalBalanceFloat = float64(totalUSD) / 100
	snapshot.CurrencyCount = len(currencyCodes)
	snapshot.CalculationDurationMs = int(time.Since(startTime).Milliseconds())

	// Save snapshot
	if err := s.snapshotRepo.CreateSnapshot(ctx, snapshot); err != nil {
		return nil, fmt.Errorf("failed to save snapshot: %w", err)
	}

	s.log.Info("snapshot_created",
		logger.String(logger.FieldUserID, userID),
		logger.String("date", date),
		logger.Float64("total_usd", snapshot.TotalBalanceFloat),
		logger.Int("account_count", snapshot.AccountCount),
		logger.Int("duration_ms", snapshot.CalculationDurationMs),
	)

	return snapshot, nil
}

// getRateAtTime finds the best rate for a currency pair at or before given timestamp.
// Returns: rate value, source, actual timestamp of rate, error
func (s *SnapshotService) getRateAtTime(ctx context.Context, from, to string, targetTime int64) (float64, string, int64, error) {
	// Try to get rate at exact time
	rate, err := s.rateRepo.GetRateAt(ctx, from, to, targetTime)
	if err == nil {
		return rate.Rate, rate.Source, rate.FetchedAt, nil
	}

	// Try reverse rate
	rate, err = s.rateRepo.GetRateAt(ctx, to, from, targetTime)
	if err == nil {
		return 1.0 / rate.Rate, rate.Source + " (inverted)", rate.FetchedAt, nil
	}

	// Try USD bridge
	if from != "USD" && to != "USD" {
		// from -> USD
		fromToUSD, err := s.rateRepo.GetRateAt(ctx, from, "USD", targetTime)
		if err != nil {
			// Try reverse
			usdToFrom, err2 := s.rateRepo.GetRateAt(ctx, "USD", from, targetTime)
			if err2 != nil {
				return 0, "", 0, fmt.Errorf("no rate found for %s to USD", from)
			}
			fromToUSD = &domain.ExchangeRate{
				Rate:      1.0 / usdToFrom.Rate,
				Source:    usdToFrom.Source,
				FetchedAt: usdToFrom.FetchedAt,
			}
		}

		// USD -> to
		if to == "USD" {
			return fromToUSD.Rate, fromToUSD.Source, fromToUSD.FetchedAt, nil
		}

		usdToTarget, err := s.rateRepo.GetRateAt(ctx, "USD", to, targetTime)
		if err != nil {
			targetToUSD, err2 := s.rateRepo.GetRateAt(ctx, to, "USD", targetTime)
			if err2 != nil {
				return 0, "", 0, fmt.Errorf("no rate found for USD to %s", to)
			}
			crossRate := fromToUSD.Rate / targetToUSD.Rate
			return crossRate, "calculated", fromToUSD.FetchedAt, nil
		}

		crossRate := fromToUSD.Rate * usdToTarget.Rate
		return crossRate, "calculated", fromToUSD.FetchedAt, nil
	}

	return 0, "", 0, fmt.Errorf("no rate available")
}

// GetHistory retrieves portfolio history for charts.
func (s *SnapshotService) GetHistory(ctx context.Context, userID string, fromDate, toDate string) ([]domain.PortfolioSnapshot, error) {
	query := domain.PortfolioHistoryQuery{
		UserID:    userID,
		FromDate:  fromDate,
		ToDate:    toDate,
		OrderDesc: false,
	}
	return s.snapshotRepo.GetHistory(ctx, query)
}

// GetSnapshot retrieves a specific day's snapshot with account details.
func (s *SnapshotService) GetSnapshot(ctx context.Context, userID string, date string) (*domain.PortfolioSnapshot, error) {
	snapshot, err := s.snapshotRepo.GetSnapshot(ctx, userID, date)
	if err != nil {
		return nil, err
	}

	accounts, err := s.snapshotRepo.GetSnapshotAccounts(ctx, snapshot.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account details: %w", err)
	}

	snapshot.Accounts = accounts
	return snapshot, nil
}

// GetTrend calculates portfolio trend over a period.
func (s *SnapshotService) GetTrend(ctx context.Context, userID string, days int) (*domain.PortfolioTrend, error) {
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	snapshots, err := s.GetHistory(ctx, userID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}

	if len(snapshots) < 2 {
		return nil, fmt.Errorf("insufficient data for trend calculation")
	}

	start := snapshots[0]
	end := snapshots[len(snapshots)-1]

	startBalance := start.TotalBalanceFloat
	endBalance := end.TotalBalanceFloat
	changeAmount := endBalance - startBalance
	changePercent := 0.0
	if startBalance > 0 {
		changePercent = (changeAmount / startBalance) * 100
	}

	// Calculate daily growth rate (CAGR-like)
	dailyGrowthRate := 0.0
	if days > 0 && startBalance > 0 {
		ratio := endBalance / startBalance
		dailyGrowthRate = (math.Pow(ratio, 1.0/float64(days)) - 1) * 100
	}

	return &domain.PortfolioTrend{
		Period:          fmt.Sprintf("%dd", days),
		StartBalance:    startBalance,
		EndBalance:      endBalance,
		ChangeAmount:    changeAmount,
		ChangePercent:   changePercent,
		DailyGrowthRate: dailyGrowthRate,
	}, nil
}
