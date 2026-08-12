package dextripay

import "time"

// Decimal is a base-10 amount represented without binary floating point.
type Decimal string

type Channel struct {
	ID                   string         `json:"id"`
	Flow                 string         `json:"flow"`
	SourceChainNamespace string         `json:"source_chain_namespace"`
	SourceChainID        string         `json:"source_chain_id"`
	SourceAsset          string         `json:"source_asset"`
	DestinationChainID   string         `json:"destination_chain_id"`
	DestinationAsset     string         `json:"destination_asset"`
	RouteFamily          string         `json:"route_family"`
	Provider             string         `json:"provider"`
	MinAmount            Decimal        `json:"min_amount,omitempty"`
	MaxAmount            Decimal        `json:"max_amount,omitempty"`
	Healthy              bool           `json:"healthy"`
	UnavailableReason    string         `json:"unavailable_reason,omitempty"`
	Metadata             map[string]any `json:"metadata,omitempty"`
}

type ListChannelsParams struct {
	Flow        string
	SourceAsset string
}

type CreateCheckoutRequest struct {
	ExternalUserID    string         `json:"external_user_id"`
	ClientReferenceID string         `json:"client_reference_id,omitempty"`
	SourceAsset       string         `json:"source_asset"`
	TargetAsset       string         `json:"target_asset"`
	Amount            Decimal        `json:"amount"`
	ReturnURL         string         `json:"return_url,omitempty"`
	Metadata          map[string]any `json:"metadata,omitempty"`
	IdempotencyKey    string         `json:"-"`
}

type checkoutCreatePayload struct {
	Type string `json:"type"`
	CreateCheckoutRequest
}

type CheckoutSession struct {
	SessionID      string         `json:"session_id"`
	OperationID    string         `json:"operation_id"`
	Type           string         `json:"type"`
	Status         string         `json:"status"`
	ExternalUserID string         `json:"external_user_id"`
	CheckoutURL    string         `json:"checkout_url"`
	QRPayload      string         `json:"qr_payload"`
	SourceAsset    string         `json:"source_asset"`
	TargetAsset    string         `json:"target_asset"`
	Amount         Decimal        `json:"amount"`
	ExpiresAt      time.Time      `json:"expires_at"`
	Metadata       map[string]any `json:"metadata,omitempty"`
}

type UserBindingSessionRequest struct {
	ExternalUserID string `json:"external_user_id"`
	WalletFamily   string `json:"wallet_family"`
	ReturnURL      string `json:"return_url,omitempty"`
	IdempotencyKey string `json:"-"`
}

type UserBindingSession struct {
	SessionID   string    `json:"session_id"`
	CheckoutURL string    `json:"checkout_url"`
	QRPayload   string    `json:"qr_payload"`
	ExpiresAt   time.Time `json:"expires_at"`
}

type Balance struct {
	Owner        string  `json:"owner"`
	SubaccountID uint64  `json:"subaccount_id"`
	Asset        string  `json:"asset"`
	Denom        string  `json:"denom"`
	Available    Decimal `json:"available"`
	Locked       Decimal `json:"locked"`
	Total        Decimal `json:"total"`
}

type UserBalances struct {
	ExternalUserID string    `json:"external_user_id"`
	Owner          string    `json:"owner"`
	Balances       []Balance `json:"balances"`
}

type ConversionMarket struct {
	AppID              string    `json:"app_id"`
	MarketID           string    `json:"market_id"`
	Pair               string    `json:"pair"`
	Enabled            bool      `json:"enabled"`
	BaseAssetID        string    `json:"base_asset_id"`
	QuoteAssetID       string    `json:"quote_asset_id"`
	BaseAssetRevision  uint64    `json:"base_asset_revision"`
	QuoteAssetRevision uint64    `json:"quote_asset_revision"`
	BaseDenom          string    `json:"base_denom"`
	QuoteDenom         string    `json:"quote_denom"`
	BaseSymbol         string    `json:"base_symbol"`
	QuoteSymbol        string    `json:"quote_symbol"`
	BaseDecimals       uint32    `json:"base_decimals"`
	QuoteDecimals      uint32    `json:"quote_decimals"`
	MarketVersion      uint64    `json:"market_version"`
	BuyBasePrice       Decimal   `json:"buy_base_price"`
	SellBasePrice      Decimal   `json:"sell_base_price"`
	PriceVersion       uint64    `json:"price_version"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type ConversionMarketList struct {
	Items []ConversionMarket `json:"items"`
}

type UpdateConversionPriceRequest struct {
	BuyBasePrice    Decimal `json:"buy_base_price"`
	SellBasePrice   Decimal `json:"sell_base_price"`
	ExpectedVersion uint64  `json:"expected_version,omitempty"`
	IdempotencyKey  string  `json:"-"`
}

type CreateConversionQuoteRequest struct {
	ExternalUserID string  `json:"external_user_id"`
	Owner          string  `json:"owner,omitempty"`
	MarketID       string  `json:"market_id"`
	Side           string  `json:"side"`
	InputAmount    Decimal `json:"input_amount"`
	IdempotencyKey string  `json:"-"`
}

type ConversionQuote struct {
	QuoteID       string    `json:"quote_id"`
	OperationID   string    `json:"operation_id"`
	MarketID      string    `json:"market_id"`
	MarketVersion uint64    `json:"market_version"`
	Pair          string    `json:"pair"`
	Side          string    `json:"side"`
	InputAsset    string    `json:"input_asset"`
	InputDenom    string    `json:"input_denom"`
	InputAmount   Decimal   `json:"input_amount"`
	OutputAsset   string    `json:"output_asset"`
	OutputDenom   string    `json:"output_denom"`
	OutputAmount  Decimal   `json:"output_amount"`
	Price         Decimal   `json:"price"`
	PriceVersion  uint64    `json:"price_version"`
	BaseAssetID   string    `json:"base_asset_id"`
	QuoteAssetID  string    `json:"quote_asset_id"`
	BaseDenom     string    `json:"base_denom"`
	QuoteDenom    string    `json:"quote_denom"`
	BaseDecimals  uint32    `json:"base_decimals"`
	QuoteDecimals uint32    `json:"quote_decimals"`
	SignDoc       string    `json:"sign_doc"`
	ExpiresAt     time.Time `json:"expires_at"`
}

type Operation struct {
	OperationID       string         `json:"operation_id"`
	ClientReferenceID string         `json:"client_reference_id,omitempty"`
	ExternalUserID    string         `json:"external_user_id"`
	Type              string         `json:"type"`
	Status            string         `json:"status"`
	SourceAsset       string         `json:"source_asset,omitempty"`
	TargetAsset       string         `json:"target_asset,omitempty"`
	InputAmount       Decimal        `json:"input_amount,omitempty"`
	OutputAmount      Decimal        `json:"output_amount,omitempty"`
	FailureCode       string         `json:"failure_code,omitempty"`
	FailureReason     string         `json:"failure_reason,omitempty"`
	Metadata          map[string]any `json:"metadata,omitempty"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
}

type ListOperationsParams struct {
	ExternalUserID string
	Type           string
	Status         string
	Cursor         string
	Limit          int
}

type OperationList struct {
	Items      []Operation `json:"items"`
	NextCursor string      `json:"next_cursor,omitempty"`
}
