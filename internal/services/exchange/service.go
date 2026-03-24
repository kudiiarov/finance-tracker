// Package exchange provides exchange rate business logic.
package exchange

import (
	"context"

	"github.com/sereozha/finance-tracker/internal/domain"
	"github.com/sereozha/finance-tracker/internal/ports/logger"
	"github.com/sereozha/finance-tracker/internal/ports/repository"
	"github.com/sereozha/finance-tracker/internal/ports/service"
)

// Service implements service.ExchangeRateService.
type Service struct {
	strategy     service.ConversionStrategy
	orchestrator *Orchestrator
	rateRepo     repository.ExchangeRateRepository
	log          logger.Logger
}

// NewService creates a new exchange rate service.
func NewService(
	currencyRepo repository.CurrencyRepository,
	rateRepo repository.ExchangeRateRepository,
	registry service.ProviderRegistry,
	log logger.Logger,
) service.ExchangeRateService {
	strategy := NewUSDBridgeStrategy(rateRepo, log)
	orchestrator := NewOrchestrator(currencyRepo, registry, rateRepo, log)

	return &Service{
		strategy:     strategy,
		orchestrator: orchestrator,
		rateRepo:     rateRepo,
		log:          log.With(logger.String(logger.FieldComponent, "exchange_service")),
	}
}

// FetchRates fetches rates from all providers using the orchestrator.
func (s *Service) FetchRates(ctx context.Context) error {
	s.log.Debug("fetch_rates_started")
	return s.orchestrator.FetchAllRates(ctx)
}

// GetRate returns the exchange rate between two currencies using the strategy.
func (s *Service) GetRate(ctx context.Context, from, to string) (*domain.ExchangeRate, error) {
	s.log.Debug("get_rate",
		logger.String(logger.FieldFromCurrency, from),
		logger.String(logger.FieldToCurrency, to),
	)

	rate, err := s.strategy.GetRate(ctx, from, to)
	if err != nil {
		s.log.Error("get_rate_failed",
			logger.Error(err),
			logger.String(logger.FieldFromCurrency, from),
			logger.String(logger.FieldToCurrency, to),
		)
		return nil, err
	}

	s.log.Debug("get_rate_success",
		logger.String(logger.FieldFromCurrency, from),
		logger.String(logger.FieldToCurrency, to),
		logger.Float64(logger.FieldRate, rate.Rate),
	)

	return rate, nil
}

// GetRateAt returns the historical rate at a specific time.
func (s *Service) GetRateAt(ctx context.Context, from, to string, at int64) (*domain.ExchangeRate, error) {
	return s.rateRepo.GetRateAt(ctx, from, to, at)
}
