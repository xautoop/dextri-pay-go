// Package checkout defines Hosted Checkout requests and session responses.
package checkout

import (
	"strings"
	"time"

	"github.com/xautoop/dextri-pay-go/api"
	"github.com/xautoop/dextri-pay-go/channels"
	"github.com/xautoop/dextri-pay-go/money"
)

// Type identifies a Hosted Checkout workflow.
type Type string

const (
	TypeDeposit           Type = "deposit"
	TypeWithdrawal        Type = "withdrawal"
	TypeConversion        Type = "conversion"
	TypeDepositAndConvert Type = "deposit_and_convert"
	TypePayment           Type = "payment"
)

// Status is the durable Hosted Checkout lifecycle state.
type Status string

const (
	StatusCreated            Status = "created"
	StatusWalletBound        Status = "wallet_bound"
	StatusAwaitingUserAction Status = "awaiting_user_action"
	StatusProcessing         Status = "processing"
	StatusSucceeded          Status = "succeeded"
	StatusFailed             Status = "failed"
	StatusExpired            Status = "expired"
)

// WalletFamily identifies the user authorization protocol family.
type WalletFamily string

const (
	WalletFamilyEVM    WalletFamily = "evm"
	WalletFamilyCosmos WalletFamily = "cosmos"
)

// CreateDepositRequest creates a stablecoin deposit checkout.
type CreateDepositRequest struct {
	// ExternalUserID is the App-side stable user identifier.
	ExternalUserID string `json:"external_user_id"`
	// ClientReferenceID is the App-side business reference.
	ClientReferenceID string `json:"client_reference_id,omitempty"`
	// SourceAsset is the external asset supplied by the user.
	SourceAsset string `json:"source_asset"`
	// TargetAsset is the final chain asset selected by the App.
	TargetAsset string `json:"target_asset"`
	// Amount is the human-readable input amount.
	Amount money.Decimal `json:"amount"`
	// ReturnURL is an App return URL previously allowlisted by Admin.
	ReturnURL string `json:"return_url,omitempty"`
	// Metadata is non-sensitive App context returned with the operation.
	Metadata api.Metadata `json:"metadata,omitempty"`
}

// Validate checks the stable deposit Checkout contract.
func (request CreateDepositRequest) Validate() error {
	return validateCreateRequest(request.ExternalUserID, request.SourceAsset, request.TargetAsset, request.Amount)
}

// CreateWithdrawalRequest creates a withdrawal checkout.
type CreateWithdrawalRequest struct {
	// ExternalUserID is the App-side stable user identifier.
	ExternalUserID string `json:"external_user_id"`
	// ClientReferenceID is the App-side business reference.
	ClientReferenceID string `json:"client_reference_id,omitempty"`
	// SourceAsset is the chain balance asset to withdraw.
	SourceAsset string `json:"source_asset"`
	// TargetAsset is the asset delivered by the selected withdrawal channel.
	TargetAsset string `json:"target_asset"`
	// Amount is the human-readable withdrawal amount.
	Amount money.Decimal `json:"amount"`
	// ReturnURL is an App return URL previously allowlisted by Admin.
	ReturnURL string `json:"return_url,omitempty"`
	// Metadata is non-sensitive App context returned with the operation.
	Metadata api.Metadata `json:"metadata,omitempty"`
}

// Validate checks the stable withdrawal Checkout contract.
func (request CreateWithdrawalRequest) Validate() error {
	return validateCreateRequest(request.ExternalUserID, request.SourceAsset, request.TargetAsset, request.Amount)
}

// CreateConversionRequest creates a chain asset-conversion checkout.
type CreateConversionRequest struct {
	// ExternalUserID is the App-side stable user identifier.
	ExternalUserID string `json:"external_user_id"`
	// ClientReferenceID is the App-side business reference.
	ClientReferenceID string `json:"client_reference_id,omitempty"`
	// SourceAsset is the configured market input asset.
	SourceAsset string `json:"source_asset"`
	// TargetAsset is the configured market output asset.
	TargetAsset string `json:"target_asset"`
	// Amount is the human-readable input amount.
	Amount money.Decimal `json:"amount"`
	// ReturnURL is an App return URL previously allowlisted by Admin.
	ReturnURL string `json:"return_url,omitempty"`
	// Metadata is non-sensitive App context returned with the operation.
	Metadata api.Metadata `json:"metadata,omitempty"`
}

// Validate checks the stable conversion Checkout contract.
func (request CreateConversionRequest) Validate() error {
	return validateCreateRequest(request.ExternalUserID, request.SourceAsset, request.TargetAsset, request.Amount)
}

// CreateDepositAndConvertRequest funds first and then requests a fresh conversion signature.
type CreateDepositAndConvertRequest struct {
	// ExternalUserID is the App-side stable user identifier.
	ExternalUserID string `json:"external_user_id"`
	// ClientReferenceID is the App-side business reference.
	ClientReferenceID string `json:"client_reference_id,omitempty"`
	// SourceAsset is the external funding asset.
	SourceAsset string `json:"source_asset"`
	// TargetAsset is the final asset after funding and conversion.
	TargetAsset string `json:"target_asset"`
	// Amount is the human-readable funding amount.
	Amount money.Decimal `json:"amount"`
	// ReturnURL is an App return URL previously allowlisted by Admin.
	ReturnURL string `json:"return_url,omitempty"`
	// Metadata is non-sensitive App context returned with the operation.
	Metadata api.Metadata `json:"metadata,omitempty"`
}

// Validate checks the stable deposit-and-convert Checkout contract.
func (request CreateDepositAndConvertRequest) Validate() error {
	return validateCreateRequest(request.ExternalUserID, request.SourceAsset, request.TargetAsset, request.Amount)
}

// CreatePaymentRequest creates a user-authorized App Commerce payment.
// Amount uses display units; Pay resolves chain decimals before authorization.
type CreatePaymentRequest struct {
	ExternalUserID    string        `json:"external_user_id"`
	ClientReferenceID string        `json:"client_reference_id"`
	Asset             string        `json:"asset"`
	Amount            money.Decimal `json:"amount"`
	ReturnURL         string        `json:"return_url,omitempty"`
	Metadata          api.Metadata  `json:"metadata,omitempty"`
}

func (request CreatePaymentRequest) Validate() error {
	return validateCreateRequest(request.ExternalUserID, request.Asset, request.Asset, request.Amount)
}

func validateCreateRequest(externalUserID, sourceAsset, targetAsset string, amount money.Decimal) error {
	if strings.TrimSpace(externalUserID) == "" {
		return &api.ValidationError{Field: "external_user_id", Message: "is required"}
	}
	if strings.TrimSpace(sourceAsset) == "" || strings.TrimSpace(targetAsset) == "" {
		return &api.ValidationError{Field: "asset", Message: "source_asset and target_asset are required"}
	}
	if err := amount.ValidatePositive(); err != nil {
		return &api.ValidationError{Field: "amount", Message: err.Error()}
	}
	return nil
}

// Session is durable Hosted Checkout state.
type Session struct {
	// SessionID identifies the durable Hosted Checkout session.
	SessionID string `json:"session_id"`
	// OperationID identifies the related durable business operation.
	OperationID string `json:"operation_id"`
	// Type identifies the Checkout workflow.
	Type Type `json:"type"`
	// Status is the latest persisted Checkout state.
	Status Status `json:"status"`
	// ExternalUserID is the App-side stable user identifier.
	ExternalUserID string `json:"external_user_id"`
	// ClientReferenceID is the App-side business reference.
	ClientReferenceID string `json:"client_reference_id,omitempty"`
	// CheckoutURL opens the hosted user flow.
	CheckoutURL string `json:"checkout_url,omitempty"`
	// QRPayload is the payload an App may render as a QR code.
	QRPayload string `json:"qr_payload,omitempty"`
	// SourceAsset is the requested input asset.
	SourceAsset string `json:"source_asset"`
	// TargetAsset is the requested final asset.
	TargetAsset string `json:"target_asset"`
	// Amount is the human-readable requested amount.
	Amount money.Decimal `json:"amount"`
	// Owner is the bound Dextri chain owner when available.
	Owner string `json:"owner,omitempty"`
	// WalletFamily is the connected wallet protocol family.
	WalletFamily WalletFamily `json:"wallet_family,omitempty"`
	// WalletAddress is the connected external wallet address.
	WalletAddress string `json:"wallet_address,omitempty"`
	// ChannelID is the authorized funding or withdrawal route.
	ChannelID string `json:"channel_id,omitempty"`
	// FundingReference identifies the underlying Funding workflow.
	FundingReference string `json:"funding_reference,omitempty"`
	// ConversionQuoteID identifies the related conversion quote.
	ConversionQuoteID string `json:"conversion_quote_id,omitempty"`
	// ReturnURL is the validated App return URL.
	ReturnURL string `json:"return_url,omitempty"`
	// Metadata is the non-sensitive context supplied by the App.
	Metadata api.Metadata `json:"metadata,omitempty"`
	// Channels contains the currently selectable authorized routes.
	Channels []channels.Channel `json:"channels,omitempty"`
	// ExpiresAt is the session expiration time.
	ExpiresAt time.Time `json:"expires_at"`
	// CreatedAt is the session creation time.
	CreatedAt time.Time `json:"created_at,omitempty"`
	// UpdatedAt is the latest persisted update time.
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}
