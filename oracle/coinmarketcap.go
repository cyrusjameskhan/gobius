package oracle

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/rs/zerolog"
)

const (
	coinmarketcapAPI = "https://pro-api.coinmarketcap.com/v2"
)

type CoinMarketCapOracle struct {
	apiKey     string
	client     *http.Client
	logger     zerolog.Logger
	priceCache map[string]priceCacheEntry
	cacheTTL   time.Duration
}

type priceCacheEntry struct {
	Price     float64
	Timestamp time.Time
}

type CMCResponse struct {
	Data map[string]struct {
		Quote struct {
			USD struct {
				Price float64 `json:"price"`
			} `json:"USD"`
		} `json:"quote"`
	} `json:"data"`
}

func NewCoinMarketCapOracle(logger zerolog.Logger, cacheTTL time.Duration) (*CoinMarketCapOracle, error) {
	apiKey := os.Getenv("COINMARKETCAP_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("COINMARKETCAP_API_KEY environment variable not set")
	}

	return &CoinMarketCapOracle{
		apiKey:     apiKey,
		client:     &http.Client{Timeout: 10 * time.Second},
		logger:     logger,
		priceCache: make(map[string]priceCacheEntry),
		cacheTTL:   cacheTTL,
	}, nil
}

// NewCoinMarketCapOracleWithConfig creates a new CoinMarketCap oracle with the given config
func NewCoinMarketCapOracleWithConfig(apiKey string, logger zerolog.Logger, cacheTTL time.Duration) (*CoinMarketCapOracle, error) {
	if apiKey == "" {
		// Try to get from environment as fallback
		apiKey = os.Getenv("COINMARKETCAP_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("CoinMarketCap API key not provided and COINMARKETCAP_API_KEY environment variable not set")
		}
	}

	return &CoinMarketCapOracle{
		apiKey:     apiKey,
		client:     &http.Client{Timeout: 10 * time.Second},
		logger:     logger,
		priceCache: make(map[string]priceCacheEntry),
		cacheTTL:   cacheTTL,
	}, nil
}

func (o *CoinMarketCapOracle) GetPrices() (basePrice float64, ethPrice float64, err error) {
	// Get ETH price
	ethPrice, err = o.getTokenPrice("ETH")
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get ETH price: %w", err)
	}

	// Get AIUS price (assuming it's listed on CMC)
	basePrice, err = o.getTokenPrice("AIUS")
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get AIUS price: %w", err)
	}

	return basePrice, ethPrice, nil
}

func (o *CoinMarketCapOracle) getTokenPrice(symbol string) (float64, error) {
	// Check cache first
	if entry, found := o.priceCache[symbol]; found {
		if time.Since(entry.Timestamp) < o.cacheTTL {
			return entry.Price, nil
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	url := fmt.Sprintf("%s/cryptocurrency/quotes/latest?symbol=%s", coinmarketcapAPI, symbol)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-CMC_PRO_API_KEY", o.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to get price: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var cmcResp CMCResponse
	if err := json.NewDecoder(resp.Body).Decode(&cmcResp); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	// Get the first (and should be only) price from the response
	for _, data := range cmcResp.Data {
		price := data.Quote.USD.Price
		// Update cache
		o.priceCache[symbol] = priceCacheEntry{
			Price:     price,
			Timestamp: time.Now(),
		}
		return price, nil
	}

	return 0, fmt.Errorf("no price data found for %s", symbol)
} 