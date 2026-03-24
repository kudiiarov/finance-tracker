package exchange

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sereozha/finance-tracker/internal/domain"
)

// CoinGeckoProvider fetches crypto exchange rates from CoinGecko API.
type CoinGeckoProvider struct {
	client  *http.Client
	baseURL string
	cap     domain.ProviderCapability
}

// coingeckoMapping maps our currency codes to CoinGecko API IDs.
// This can be extended or loaded from config.
var coingeckoMapping = map[string]string{
	"BTC":    "bitcoin",
	"ETH":    "ethereum",
	"SOL":    "solana",
	"ADA":    "cardano",
	"DOT":    "polkadot",
	"LINK":   "chainlink",
	"UNI":    "uniswap",
	"AAVE":   "aave",
	"SUSHI":  "sushi",
	"COMP":   "compound-governance-token",
	"MKR":    "maker",
	"YFI":    "yearn-finance",
	"SNX":    "havven",
	"CRV":    "curve-dao-token",
	"1INCH":  "1inch",
	"GRT":    "the-graph",
	"LDO":    "lido-dao",
	"STETH":  "staked-ether",
	"MATIC":  "matic-network",
	"AVAX":   "avalanche-2",
	"FTM":    "fantom",
	"ARB":    "arbitrum",
	"OP":     "optimism",
	"BNB":    "binancecoin",
	"XRP":    "ripple",
	"DOGE":   "dogecoin",
	"TRX":    "tron",
	"LTC":    "litecoin",
	"BCH":    "bitcoin-cash",
	"ETC":    "ethereum-classic",
	"XLM":    "stellar",
	"XMR":    "monero",
	"ALGO":   "algorand",
	"VET":    "vechain",
	"FIL":    "filecoin",
	"EOS":    "eos",
	"XTZ":    "tezos",
	"ATOM":   "cosmos",
	"NEAR":   "near",
	"ICP":    "internet-computer",
	"APT":    "aptos",
	"SUI":    "sui",
	"SEI":    "sei-network",
	"TIA":    "celestia",
	"DYM":    "dymension",
	"STRK":   "starknet",
	"ZKS":    "zksync",
	"BASE":   "base",
	"SCROLL": "scroll",
	"MANTA":  "manta-network",
}

// NewCoinGeckoProvider creates a new CoinGecko provider.
func NewCoinGeckoProvider() *CoinGeckoProvider {
	return &CoinGeckoProvider{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://api.coingecko.com/api/v3",
		cap: domain.ProviderCapability{
			Name:           "coingecko",
			Types:          []domain.CurrencyType{domain.CurrencyTypeCrypto},
			BaseCurrencies: []string{"USD"},
			Priority:       1,
			RateLimit: domain.RateLimitConfig{
				RequestsPerMinute: 50,
				RequestsPerHour:   100,
			},
		},
	}
}

// GetCapabilities returns what this provider can fetch.
func (p *CoinGeckoProvider) GetCapabilities() domain.ProviderCapability {
	return p.cap
}

// FetchRates fetches crypto rates from CoinGecko.
func (p *CoinGeckoProvider) FetchRates(ctx context.Context, request domain.RateRequest) ([]domain.ExchangeRate, error) {
	// Build list of CoinGecko IDs from supported currencies
	var ids []string
	codeToID := make(map[string]string) // reverse mapping for results

	// If specific currencies requested, only fetch those
	if len(request.Currencies) > 0 {
		for _, code := range request.Currencies {
			if id, ok := coingeckoMapping[code]; ok {
				ids = append(ids, id)
				codeToID[id] = code
			}
		}
	} else {
		// Fetch all known currencies
		for code, id := range coingeckoMapping {
			ids = append(ids, id)
			codeToID[id] = code
		}
	}

	if len(ids) == 0 {
		return []domain.ExchangeRate{}, nil
	}

	url := fmt.Sprintf("%s/simple/price?ids=%s&vs_currencies=usd", p.baseURL, strings.Join(ids, ","))

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var result map[string]map[string]float64
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	now := time.Now().Unix()
	rates := make([]domain.ExchangeRate, 0)

	for coinID, data := range result {
		if rate, ok := data["usd"]; ok {
			if currencyCode, ok := codeToID[coinID]; ok {
				rates = append(rates, domain.ExchangeRate{
					ID:           uuid.New(),
					FromCurrency: currencyCode,
					ToCurrency:   "USD",
					Rate:         rate,
					Source:       p.cap.Name,
					FetchedAt:    now,
				})
			}
		}
	}

	return rates, nil
}

// GetCoinGeckoID returns the CoinGecko ID for a currency code.
func GetCoinGeckoID(code string) (string, bool) {
	id, ok := coingeckoMapping[code]
	return id, ok
}

// RegisterCoinGeckoID adds a new currency mapping dynamically.
func RegisterCoinGeckoID(code, id string) {
	coingeckoMapping[code] = id
}
