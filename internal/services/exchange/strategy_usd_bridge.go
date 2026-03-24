package exchange

import (
	"context"
	"fmt"

	"github.com/sereozha/finance-tracker/internal/domain"
	"github.com/sereozha/finance-tracker/internal/ports/logger"
	"github.com/sereozha/finance-tracker/internal/ports/service"
)

const (
	strategyName     = "usd_bridge"
	baseCurrency     = "USD"
	sourceCalculated = "calculated"
	sourceInverted   = "inverted"
)

// USDBridgeStrategy implements conversion via USD as intermediary.
type USDBridgeStrategy struct {
	rateRepo service.RateRepository
	log      logger.Logger
}

// NewUSDBridgeStrategy creates a new USD bridge strategy.
func NewUSDBridgeStrategy(rateRepo service.RateRepository, log logger.Logger) service.ConversionStrategy {
	return &USDBridgeStrategy{
		rateRepo: rateRepo,
		log:      log.With(logger.String(logger.FieldComponent, strategyName)),
	}
}

// Name returns the strategy identifier.
func (s *USDBridgeStrategy) Name() string {
	return strategyName
}

// CanConvert checks if this strategy can convert the given pair.
func (s *USDBridgeStrategy) CanConvert(ctx context.Context, from, to string) bool {
	if from == baseCurrency || to == baseCurrency {
		return s.hasRateToUSD(ctx, from) || s.hasRateToUSD(ctx, to)
	}
	return s.hasRateToUSD(ctx, from) && s.hasRateToUSD(ctx, to)
}

// GetRate returns the exchange rate between two currencies.
func (s *USDBridgeStrategy) GetRate(ctx context.Context, from, to string) (*domain.ExchangeRate, error) {
	s.log.Debug("calculating_rate",
		logger.String(logger.FieldFromCurrency, from),
		logger.String(logger.FieldToCurrency, to),
	)

	// Try direct rate first
	if rate, err := s.getDirectRate(ctx, from, to); err == nil {
		return rate, nil
	}

	// Try reverse rate
	if rate, err := s.getReverseRate(ctx, from, to); err == nil {
		return rate, nil
	}

	// Fall back to USD bridge
	return s.getCrossRateViaUSD(ctx, from, to)
}

// getDirectRate tries to get a direct rate from repository.
func (s *USDBridgeStrategy) getDirectRate(ctx context.Context, from, to string) (*domain.ExchangeRate, error) {
	rate, err := s.rateRepo.GetLatest(ctx, from, to)
	if err == nil {
		s.log.Debug("direct_rate_found",
			logger.String(logger.FieldFromCurrency, from),
			logger.String(logger.FieldToCurrency, to),
		)
	}
	return rate, err
}

// getReverseRate tries to get rate by inverting the reverse pair.
func (s *USDBridgeStrategy) getReverseRate(ctx context.Context, from, to string) (*domain.ExchangeRate, error) {
	rate, err := s.rateRepo.GetLatest(ctx, to, from)
	if err != nil {
		return nil, err
	}

	s.log.Debug("reverse_rate_used",
		logger.String(logger.FieldFromCurrency, from),
		logger.String(logger.FieldToCurrency, to),
	)

	return &domain.ExchangeRate{
		FromCurrency: from,
		ToCurrency:   to,
		Rate:         1.0 / rate.Rate,
		Source:       fmt.Sprintf("%s (%s)", rate.Source, sourceInverted),
		FetchedAt:    rate.FetchedAt,
	}, nil
}

// getCrossRateViaUSD calculates rate via USD intermediary.
func (s *USDBridgeStrategy) getCrossRateViaUSD(ctx context.Context, from, to string) (*domain.ExchangeRate, error) {
	s.log.Debug("calculating_via_usd_bridge",
		logger.String(logger.FieldFromCurrency, from),
		logger.String(logger.FieldToCurrency, to),
	)

	fromRate, err := s.getRateToUSD(ctx, from)
	if err != nil {
		return nil, fmt.Errorf("no rate available for %s to %s: %w", from, baseCurrency, err)
	}

	// If target is USD, return the rate we already have
	if to == baseCurrency {
		return fromRate, nil
	}

	toRate, err := s.getRateFromUSD(ctx, to)
	if err != nil {
		return nil, fmt.Errorf("no rate available for %s to %s: %w", baseCurrency, to, err)
	}

	crossRate := fromRate.Rate * toRate.Rate

	s.log.Debug("usd_bridge_calculation",
		logger.String(logger.FieldFromCurrency, from),
		logger.String(logger.FieldToCurrency, to),
		logger.Float64("cross_rate", crossRate),
	)

	return &domain.ExchangeRate{
		FromCurrency: from,
		ToCurrency:   to,
		Rate:         crossRate,
		Source:       sourceCalculated,
		FetchedAt:    fromRate.FetchedAt,
	}, nil
}

// hasRateToUSD checks if a rate exists between currency and USD.
func (s *USDBridgeStrategy) hasRateToUSD(ctx context.Context, currency string) bool {
	if currency == baseCurrency {
		return true
	}
	_, err1 := s.rateRepo.GetLatest(ctx, currency, baseCurrency)
	_, err2 := s.rateRepo.GetLatest(ctx, baseCurrency, currency)
	return err1 == nil || err2 == nil
}

// getRateToUSD gets the rate from currency to USD.
func (s *USDBridgeStrategy) getRateToUSD(ctx context.Context, currency string) (*domain.ExchangeRate, error) {
	if currency == baseCurrency {
		return &domain.ExchangeRate{
			FromCurrency: currency,
			ToCurrency:   baseCurrency,
			Rate:         1.0,
			Source:       "fixed",
		}, nil
	}

	// Try direct rate
	if rate, err := s.rateRepo.GetLatest(ctx, currency, baseCurrency); err == nil {
		return rate, nil
	}

	// Try reverse and invert
	return s.getReverseRate(ctx, currency, baseCurrency)
}

// getRateFromUSD gets the rate from USD to currency.
func (s *USDBridgeStrategy) getRateFromUSD(ctx context.Context, currency string) (*domain.ExchangeRate, error) {
	if currency == baseCurrency {
		return &domain.ExchangeRate{
			FromCurrency: baseCurrency,
			ToCurrency:   currency,
			Rate:         1.0,
			Source:       "fixed",
		}, nil
	}

	// Try direct rate
	if rate, err := s.rateRepo.GetLatest(ctx, baseCurrency, currency); err == nil {
		return rate, nil
	}

	// Try reverse and invert
	return s.getReverseRate(ctx, baseCurrency, currency)
}
