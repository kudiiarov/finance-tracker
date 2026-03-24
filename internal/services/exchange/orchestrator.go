package exchange

import (
	"context"
	"fmt"
	"sync"

	"github.com/sereozha/finance-tracker/internal/domain"
	"github.com/sereozha/finance-tracker/internal/ports/logger"
	"github.com/sereozha/finance-tracker/internal/ports/repository"
	"github.com/sereozha/finance-tracker/internal/ports/service"
)

// Orchestrator coordinates rate fetching with fallback support.
type Orchestrator struct {
	currencyRepo repository.CurrencyRepository
	registry     service.ProviderRegistry
	rateRepo     repository.ExchangeRateRepository
	log          logger.Logger
}

// NewOrchestrator creates a new orchestrator.
func NewOrchestrator(
	currencyRepo repository.CurrencyRepository,
	registry service.ProviderRegistry,
	rateRepo repository.ExchangeRateRepository,
	log logger.Logger,
) *Orchestrator {
	return &Orchestrator{
		currencyRepo: currencyRepo,
		registry:     registry,
		rateRepo:     rateRepo,
		log:          log.With(logger.String(logger.FieldComponent, "exchange_orchestrator")),
	}
}

// FetchAllRates fetches rates for all active currencies using their configured providers.
func (o *Orchestrator) FetchAllRates(ctx context.Context) error {
	// Get all active currencies
	currencies, err := o.currencyRepo.ListActive(ctx)
	if err != nil {
		return fmt.Errorf("failed to list currencies: %w", err)
	}

	o.log.Debug("fetching_rates_for_currencies",
		logger.Int("count", len(currencies)),
	)

	// Group currencies by their primary provider
	byProvider := make(map[string][]domain.CurrencyInfo)
	for _, c := range currencies {
		primary := c.GetPrimaryProvider()
		if primary != "" {
			byProvider[primary] = append(byProvider[primary], c)
		}
	}

	o.log.Debug("grouped_by_provider",
		logger.Int("provider_count", len(byProvider)),
	)

	// Fetch from each provider concurrently
	var wg sync.WaitGroup
	errChan := make(chan error, len(byProvider))

	for providerName, curs := range byProvider {
		wg.Add(1)
		go func(name string, currencies []domain.CurrencyInfo) {
			defer wg.Done()
			if err := o.fetchFromProvider(ctx, name, currencies); err != nil {
				errChan <- fmt.Errorf("provider %s: %w", name, err)
			}
		}(providerName, curs)
	}

	wg.Wait()
	close(errChan)

	// Collect errors
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		o.log.Error("some_providers_failed",
			logger.Int("error_count", len(errs)),
		)
		return fmt.Errorf("some providers failed: %v", errs)
	}

	o.log.Info("rate_fetch_completed")
	return nil
}

// fetchFromProvider fetches rates from a specific provider with fallback.
func (o *Orchestrator) fetchFromProvider(ctx context.Context, providerName string, currencies []domain.CurrencyInfo) error {
	o.log.Debug("fetching_from_provider",
		logger.String(logger.FieldProvider, providerName),
		logger.Int("currency_count", len(currencies)),
	)

	// Try primary provider
	provider, ok := o.registry.GetProvider(providerName)
	if !ok {
		o.log.Error("provider_not_found_in_registry",
			logger.String(logger.FieldProvider, providerName),
		)
		return fmt.Errorf("provider %s not found in registry", providerName)
	}

	// Build list of currency codes to fetch
	currencyCodes := make([]string, len(currencies))
	for i, c := range currencies {
		currencyCodes[i] = c.Code
	}

	request := domain.RateRequest{
		PreferredProviders: []string{providerName},
		Currencies:         currencyCodes,
	}

	rates, err := provider.FetchRates(ctx, request)
	if err == nil && len(rates) > 0 {
		o.log.Debug("primary_provider_success",
			logger.String(logger.FieldProvider, providerName),
			logger.Int("rate_count", len(rates)),
		)
		return o.saveRates(ctx, rates)
	}

	// Primary failed, try fallbacks
	if err != nil {
		o.log.Warn("primary_provider_failed",
			logger.String(logger.FieldProvider, providerName),
			logger.Error(err),
		)
	} else {
		o.log.Warn("primary_provider_empty_result",
			logger.String(logger.FieldProvider, providerName),
			logger.String("reason", "provider returned no rates"),
			logger.Strings("requested_currencies", currencyCodes),
		)
	}

	for _, currency := range currencies {
		fallbacks := currency.GetFallbackProviders()
		for _, fallbackName := range fallbacks {
			fallback, ok := o.registry.GetProvider(fallbackName)
			if !ok {
				o.log.Debug("fallback_provider_not_found",
					logger.String(logger.FieldProvider, fallbackName),
				)
				continue
			}

			o.log.Debug("trying_fallback_provider",
				logger.String("currency", currency.Code),
				logger.String("fallback", fallbackName),
			)

			fallbackReq := domain.RateRequest{
				PreferredProviders: []string{fallbackName},
			}

			rates, err := fallback.FetchRates(ctx, fallbackReq)
			if err == nil && len(rates) > 0 {
				// Only save rates for the requested currency
				var filtered []domain.ExchangeRate
				for _, r := range rates {
					if r.FromCurrency == currency.Code {
						filtered = append(filtered, r)
					}
				}
				if err := o.saveRates(ctx, filtered); err != nil {
					o.log.Error("failed_to_save_fallback_rates",
						logger.Error(err),
						logger.String("currency", currency.Code),
					)
				} else {
					o.log.Info("fallback_provider_success",
						logger.String("currency", currency.Code),
						logger.String("provider", fallbackName),
					)
				}
				break // Success, no more fallbacks needed
			}
		}
	}

	return nil
}

func (o *Orchestrator) saveRates(ctx context.Context, rates []domain.ExchangeRate) error {
	for _, rate := range rates {
		if err := o.rateRepo.Save(ctx, &rate); err != nil {
			o.log.Error("failed_to_save_rate",
				logger.Error(err),
				logger.String(logger.FieldFromCurrency, rate.FromCurrency),
				logger.String(logger.FieldToCurrency, rate.ToCurrency),
			)
		}
	}
	return nil
}
