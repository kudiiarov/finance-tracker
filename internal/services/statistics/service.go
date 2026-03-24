// Package statistics provides financial statistics business logic.
package statistics

import (
	"context"
	"fmt"

	"github.com/sereozha/finance-tracker/internal/domain"
	"github.com/sereozha/finance-tracker/internal/ports/repository"
	"github.com/sereozha/finance-tracker/internal/ports/service"
)

// Service implements service.StatisticsService.
type Service struct {
	accountRepo repository.AccountRepository
	rateSvc     service.ExchangeRateService
	settingsSvc service.SettingsService
}

// NewService creates a new statistics service.
func NewService(
	accountRepo repository.AccountRepository,
	rateSvc service.ExchangeRateService,
	settingsSvc service.SettingsService,
) service.StatisticsService {
	return &Service{
		accountRepo: accountRepo,
		rateSvc:     rateSvc,
		settingsSvc: settingsSvc,
	}
}

// GetTotalBalance calculates total wealth in target currency.
func (s *Service) GetTotalBalance(ctx context.Context, userID string, targetCurrency string) (*service.StatisticsResult, error) {
	accounts, err := s.accountRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch accounts: %w", err)
	}

	if targetCurrency == "" {
		targetCurrency, _ = s.settingsSvc.GetMainCurrency(ctx, userID)
	}

	targetCurrencyObj, ok := domain.GetCurrency(targetCurrency)
	if !ok {
		return nil, fmt.Errorf("invalid target currency: %s", targetCurrency)
	}

	var totalAmount int64
	ratesUsed := make([]service.RateInfo, 0)

	for _, account := range accounts {
		if account.Currency.Code == targetCurrency {
			totalAmount += account.Amount
		} else {
			rate, err := s.rateSvc.GetRate(ctx, account.Currency.Code, targetCurrency)
			if err != nil {
				continue
			}

			converted := rate.Convert(account.Amount, account.Currency.Precision, targetCurrencyObj.Precision)
			totalAmount += converted

			// Track rate used (avoid duplicates)
			found := false
			for _, r := range ratesUsed {
				if r.FromCurrency == account.Currency.Code && r.ToCurrency == targetCurrency {
					found = true
					break
				}
			}
			if !found {
				ratesUsed = append(ratesUsed, service.RateInfo{
					FromCurrency: account.Currency.Code,
					ToCurrency:   targetCurrency,
					Rate:         rate.Rate,
					Source:       rate.Source,
				})
			}
		}
	}

	return &service.StatisticsResult{
		UserID:         userID,
		TargetCurrency: targetCurrency,
		TotalAmount:    totalAmount,
		TotalFloat:     targetCurrencyObj.FromSmallestUnit(totalAmount),
		Accounts:       len(accounts),
		RatesUsed:      ratesUsed,
	}, nil
}
