// Package conversion defines App-priced, chain-backed market contracts.
package conversion

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/xautoop/dextri-pay-go/api"
	"github.com/xautoop/dextri-pay-go/money"
)

// Side identifies which configured market asset the user buys or sells.
type Side string

const (
	SideSellBase Side = "sell_base"
	SideBuyBase  Side = "buy_base"

	// MaxPriceScale is the Partner API price-field scale. It is independent of
	// the base and quote asset precisions returned by the chain asset registry.
	MaxPriceScale = 18
)

// Valid reports whether side is supported by the conversion API.
func (side Side) Valid() bool { return side == SideSellBase || side == SideBuyBase }

// Market is an Admin-configured App market backed by active chain assets.
type Market struct {
	// AppID identifies the App that owns the market and reserves.
	AppID string `json:"app_id"`
	// MarketID is the stable Admin-assigned market identifier.
	MarketID string `json:"market_id"`
	// Pair is the display pair derived from the configured assets.
	Pair string `json:"pair"`
	// Enabled reports whether new quotes may be created.
	Enabled bool `json:"enabled"`
	// BaseAssetID is the base asset identifier from the chain registry.
	BaseAssetID string `json:"base_asset_id"`
	// QuoteAssetID is the quote asset identifier from the chain registry.
	QuoteAssetID string `json:"quote_asset_id"`
	// BaseAssetRevision snapshots the active base asset configuration.
	BaseAssetRevision uint64 `json:"base_asset_revision"`
	// QuoteAssetRevision snapshots the active quote asset configuration.
	QuoteAssetRevision uint64 `json:"quote_asset_revision"`
	// BaseDenom is the base asset atomic chain denomination.
	BaseDenom string `json:"base_denom"`
	// QuoteDenom is the quote asset atomic chain denomination.
	QuoteDenom string `json:"quote_denom"`
	// BaseSymbol is the base asset display symbol.
	BaseSymbol string `json:"base_symbol"`
	// QuoteSymbol is the quote asset display symbol.
	QuoteSymbol string `json:"quote_symbol"`
	// BaseDecimals is the chain-owned base asset precision snapshot.
	BaseDecimals uint32 `json:"base_decimals"`
	// QuoteDecimals is the chain-owned quote asset precision snapshot.
	QuoteDecimals uint32 `json:"quote_decimals"`
	// MarketVersion increases when the configured asset pair changes.
	MarketVersion uint64 `json:"market_version"`
	// BuyBasePrice is the quote amount the App pays for one base unit.
	BuyBasePrice money.Decimal `json:"buy_base_price"`
	// SellBasePrice is the quote amount the App charges for one base unit.
	SellBasePrice money.Decimal `json:"sell_base_price"`
	// PriceVersion increases whenever an App price changes.
	PriceVersion uint64 `json:"price_version"`
	// UpdatedAt is the latest persisted market update time.
	UpdatedAt time.Time `json:"updated_at"`
}

// MarketList contains all conversion markets visible to the authenticated App.
type MarketList struct {
	// Items contains Admin-configured markets.
	Items []Market `json:"items"`
}

// UpdatePriceRequest replaces both sides of an App-defined price.
type UpdatePriceRequest struct {
	// BuyBasePrice is the quote amount paid when the App buys one base unit.
	BuyBasePrice money.Decimal `json:"buy_base_price"`
	// SellBasePrice is the quote amount charged when the App sells one base unit.
	SellBasePrice money.Decimal `json:"sell_base_price"`
	// ExpectedVersion is the optional optimistic concurrency version.
	ExpectedVersion uint64 `json:"expected_version,omitempty"`
}

// Validate checks the complete Partner API price contract without applying
// either asset's chain precision.
func (request UpdatePriceRequest) Validate() error {
	if err := request.BuyBasePrice.ValidatePositive(); err != nil {
		return fmt.Errorf("buy_base_price: %w", err)
	}
	if err := request.SellBasePrice.ValidatePositive(); err != nil {
		return fmt.Errorf("sell_base_price: %w", err)
	}
	if decimalScale(request.BuyBasePrice) > MaxPriceScale || decimalScale(request.SellBasePrice) > MaxPriceScale {
		return fmt.Errorf("prices must contain at most %d decimal places", MaxPriceScale)
	}
	buy, _ := new(big.Rat).SetString(request.BuyBasePrice.String())
	sell, _ := new(big.Rat).SetString(request.SellBasePrice.String())
	if buy.Cmp(sell) > 0 {
		return fmt.Errorf("buy_base_price must not exceed sell_base_price")
	}
	return nil
}

func decimalScale(value money.Decimal) int {
	raw := value.String()
	separator := strings.IndexByte(raw, '.')
	if separator < 0 {
		return 0
	}
	return len(raw) - separator - 1
}

// CreateQuoteRequest creates a short-lived user-signable quote.
type CreateQuoteRequest struct {
	// ExternalUserID is the App-side stable user identifier.
	ExternalUserID string `json:"external_user_id"`
	// Owner is the bound Dextri owner when one wallet must be selected.
	Owner string `json:"owner,omitempty"`
	// MarketID selects an Admin-configured active market.
	MarketID string `json:"market_id"`
	// Side selects whether the user sells or buys the configured base asset.
	Side Side `json:"side"`
	// InputAmount is the human-readable amount of the selected input asset.
	InputAmount money.Decimal `json:"input_amount"`
}

// Validate checks the stable quote-creation contract.
func (request CreateQuoteRequest) Validate() error {
	if strings.TrimSpace(request.ExternalUserID) == "" {
		return &api.ValidationError{Field: "external_user_id", Message: "is required"}
	}
	if strings.TrimSpace(request.MarketID) == "" {
		return &api.ValidationError{Field: "market_id", Message: "is required"}
	}
	if !request.Side.Valid() {
		return &api.ValidationError{Field: "side", Message: "must be sell_base or buy_base"}
	}
	if err := request.InputAmount.ValidatePositive(); err != nil {
		return &api.ValidationError{Field: "input_amount", Message: err.Error()}
	}
	return nil
}

// Quote is calculated with the App price and chain-owned asset precision.
type Quote struct {
	// QuoteID identifies the immutable short-lived quote.
	QuoteID string `json:"quote_id"`
	// OperationID identifies the related durable business operation.
	OperationID string `json:"operation_id"`
	// MarketID identifies the Admin-configured market.
	MarketID string `json:"market_id"`
	// MarketVersion snapshots the configured asset pair.
	MarketVersion uint64 `json:"market_version"`
	// Pair is the display pair derived from the configured assets.
	Pair string `json:"pair"`
	// Side specifies whether the user sells or buys the base asset.
	Side Side `json:"side"`
	// InputAsset is the chain asset identifier debited from the user.
	InputAsset string `json:"input_asset"`
	// InputDenom is the atomic chain denomination debited from the user.
	InputDenom string `json:"input_denom"`
	// InputAmount is the human-readable user input amount.
	InputAmount money.Decimal `json:"input_amount"`
	// OutputAsset is the chain asset identifier credited to the user.
	OutputAsset string `json:"output_asset"`
	// OutputDenom is the atomic chain denomination credited to the user.
	OutputDenom string `json:"output_denom"`
	// OutputAmount is the rounded-down human-readable output amount.
	OutputAmount money.Decimal `json:"output_amount"`
	// Price is the App-defined price used by this quote.
	Price money.Decimal `json:"price"`
	// PriceVersion snapshots the App price version.
	PriceVersion uint64 `json:"price_version"`
	// BaseAssetID is the configured base asset identifier.
	BaseAssetID string `json:"base_asset_id"`
	// QuoteAssetID is the configured quote asset identifier.
	QuoteAssetID string `json:"quote_asset_id"`
	// BaseDenom is the configured base asset atomic denomination.
	BaseDenom string `json:"base_denom"`
	// QuoteDenom is the configured quote asset atomic denomination.
	QuoteDenom string `json:"quote_denom"`
	// BaseDecimals is the chain-owned base asset precision snapshot.
	BaseDecimals uint32 `json:"base_decimals"`
	// QuoteDecimals is the chain-owned quote asset precision snapshot.
	QuoteDecimals uint32 `json:"quote_decimals"`
	// SignDoc is the canonical document the user authorizes.
	SignDoc string `json:"sign_doc"`
	// SignPayload contains wallet-family-specific signing data.
	SignPayload json.RawMessage `json:"sign_payload,omitempty"`
	// ExpiresAt is the last time at which the quote may be authorized.
	ExpiresAt time.Time `json:"expires_at"`
	// SignatureRequired reports whether a user wallet signature is required.
	SignatureRequired bool `json:"signature_required"`
	// InteractionType is none for App-custodial conversion and wallet_signature otherwise.
	InteractionType string `json:"interaction_type"`
}
