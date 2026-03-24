package exchange

import (
	"sync"

	"github.com/sereozha/finance-tracker/internal/domain"
	"github.com/sereozha/finance-tracker/internal/ports/service"
)

// CapabilityRegistry manages capability-based providers.
type CapabilityRegistry struct {
	providers map[string]service.CapabilityProvider
	mu        sync.RWMutex
}

// NewCapabilityRegistry creates a new registry with default providers.
func NewCapabilityRegistry() *CapabilityRegistry {
	r := &CapabilityRegistry{
		providers: make(map[string]service.CapabilityProvider),
	}

	// Register default providers
	r.Register(NewCoinGeckoProvider())
	r.Register(NewExchangeRateAPIProvider())
	r.Register(NewMetalsAPIProvider())

	return r
}

// Register adds a provider to the registry.
func (r *CapabilityRegistry) Register(provider service.CapabilityProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()

	cap := provider.GetCapabilities()
	r.providers[cap.Name] = provider
}

// GetProvider returns a provider by name.
func (r *CapabilityRegistry) GetProvider(name string) (service.CapabilityProvider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.providers[name]
	return p, ok
}

// FindProvidersFor returns providers that can handle a currency type.
func (r *CapabilityRegistry) FindProvidersFor(currencyType domain.CurrencyType) []service.CapabilityProvider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []service.CapabilityProvider
	for _, p := range r.providers {
		cap := p.GetCapabilities()
		for _, t := range cap.Types {
			if t == currencyType {
				result = append(result, p)
				break
			}
		}
	}
	return result
}

// FindProvidersForPair returns providers that can fetch a specific pair.
func (r *CapabilityRegistry) FindProvidersForPair(from, to string) []service.CapabilityProvider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// For now, we return all providers - they will filter internally
	// In a more sophisticated system, we'd check if provider supports specific pair
	var result []service.CapabilityProvider
	for _, p := range r.providers {
		result = append(result, p)
	}
	return result
}

// GetAllProviders returns all registered providers.
func (r *CapabilityRegistry) GetAllProviders() []service.CapabilityProvider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []service.CapabilityProvider
	for _, p := range r.providers {
		result = append(result, p)
	}
	return result
}
