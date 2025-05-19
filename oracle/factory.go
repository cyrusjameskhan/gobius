package oracle

import (
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"gobius/config"
)

// PriceOracle defines the interface for getting token prices
type PriceOracle interface {
	GetPrices() (basePrice float64, ethPrice float64, err error)
}

// NewPriceOracle creates a new price oracle based on the configuration
func NewPriceOracle(cfg *config.AppConfig, logger zerolog.Logger) (PriceOracle, error) {
	cacheTTL, err := time.ParseDuration(cfg.PriceOracleCacheTTL)
	if err != nil {
		return nil, fmt.Errorf("invalid price oracle cache TTL: %w", err)
	}

	switch cfg.PriceOracleType {
	case "paraswap":
		timeout, err := time.ParseDuration(cfg.ParaswapTimeout)
		if err != nil {
			return nil, fmt.Errorf("invalid paraswap timeout: %w", err)
		}
		return NewParaswapOracle(logger, cacheTTL, timeout)
	case "coinmarketcap":
		return NewCoinMarketCapOracleWithConfig(cfg.CoinMarketCapAPIKey, logger, cacheTTL)
	default:
		return nil, fmt.Errorf("unsupported price oracle type: %s", cfg.PriceOracleType)
	}
} 