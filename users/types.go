// Package users defines App-user wallet binding and balance contracts.
package users

import (
	"strings"
	"time"

	"github.com/xautoop/dextri-pay-go/api"
	"github.com/xautoop/dextri-pay-go/money"
)

// Balance is one chain-backed balance for a bound owner.
type Balance struct {
	// Owner is the Dextri account that owns the balance.
	Owner string `json:"owner"`
	// SubaccountID is the Vault subaccount identifier.
	SubaccountID uint64 `json:"subaccount_id"`
	// Asset is the active chain asset identifier.
	Asset string `json:"asset"`
	// Denom is the atomic chain denomination.
	Denom string `json:"denom"`
	// Available is the unlocked human-readable balance.
	Available money.Decimal `json:"available"`
	// Locked is the locked human-readable balance.
	Locked money.Decimal `json:"locked"`
	// Total is the available plus locked balance.
	Total money.Decimal `json:"total"`
}

// Balances contains live Funding/chain balances for one App user.
type Balances struct {
	// ExternalUserID is the App-side stable user identifier.
	ExternalUserID string `json:"external_user_id"`
	// Owner optionally identifies the requested bound Dextri owner.
	Owner string `json:"owner,omitempty"`
	// Items contains live Funding or chain-backed balances.
	Items []Balance `json:"balances"`
}

// WalletFamily identifies the user authorization protocol family.
type WalletFamily string

const (
	WalletFamilyEVM    WalletFamily = "evm"
	WalletFamilyCosmos WalletFamily = "cosmos"
)

// CreateBindingSessionRequest creates a wallet-binding-only Checkout session.
type CreateBindingSessionRequest struct {
	// ExternalUserID is the App-side stable user identifier.
	ExternalUserID string `json:"external_user_id"`
	// WalletFamily is the wallet protocol family the App expects.
	WalletFamily WalletFamily `json:"wallet_family"`
	// ReturnURL is an App return URL previously allowlisted by Admin.
	ReturnURL string `json:"return_url,omitempty"`
}

// Validate checks the stable wallet-binding contract.
func (request CreateBindingSessionRequest) Validate() error {
	if strings.TrimSpace(request.ExternalUserID) == "" {
		return &api.ValidationError{Field: "external_user_id", Message: "is required"}
	}
	if request.WalletFamily != WalletFamilyEVM && request.WalletFamily != WalletFamilyCosmos {
		return &api.ValidationError{Field: "wallet_family", Message: "must be evm or cosmos"}
	}
	return nil
}

// BindingSession contains a wallet-binding Checkout URL.
type BindingSession struct {
	// SessionID identifies the wallet-binding session.
	SessionID string `json:"session_id"`
	// OperationID identifies the durable wallet-binding operation.
	OperationID string `json:"operation_id"`
	// CheckoutURL opens the hosted wallet-binding flow.
	CheckoutURL string `json:"checkout_url"`
	// QRPayload is the payload an App may render as a QR code.
	QRPayload string `json:"qr_payload"`
	// ExpiresAt is the session expiration time.
	ExpiresAt time.Time `json:"expires_at"`
}
