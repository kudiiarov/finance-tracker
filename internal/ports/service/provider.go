package service

import (
	"context"

	"github.com/sereozha/finance-tracker/internal/domain"
)

// CapabilityProvider is an exchange rate provider with capability metadata.
type CapabilityProvider interface {
	// GetCapabilities returns what this provider can fetch.
	GetCapabilities() domain.ProviderCapability

	// FetchRates fetches rates for the requested pairs.
	// Only returns rates for pairs the provider supports.
	FetchRates(ctx context.Context, request domain.RateRequest) ([]domain.ExchangeRate, error)
}

// ProviderRegistry manages capability-based providers.
type ProviderRegistry interface {
	// Register adds a provider to the registry.
	Register(provider CapabilityProvider)

	// GetProvider returns a provider by name.
	GetProvider(name string) (CapabilityProvider, bool)

	// FindProvidersFor returns providers that can handle a currency type.
	FindProvidersFor(currencyType domain.CurrencyType) []CapabilityProvider

	// FindProvidersForPair returns providers that can fetch a specific pair.
	FindProvidersForPair(from, to string) []CapabilityProvider

	// GetAllProviders returns all registered providers.
	GetAllProviders() []CapabilityProvider
}
