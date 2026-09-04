// Package account defines App-scoped settlement-account contracts.
package account

import "github.com/xautoop/dextri-pay-go/money"

// Capability is one operation an App account is allowed to perform.
type Capability string

const (
	CapabilityReceive Capability = "receive"
	CapabilityPayout  Capability = "payout"
	CapabilityRefund  Capability = "refund"
	CapabilityBurn    Capability = "burn"
)

// Balance is the authoritative balance of one asset in an App account.
type Balance struct {
	AccountKey   string        `json:"account_key"`  // App-scoped account key; never an owner or subaccount ID.
	Asset        string        `json:"asset"`        // Asset code resolved by Pay.
	Total        money.Decimal `json:"total"`        // Total account balance.
	Available    money.Decimal `json:"available"`    // Amount currently available for authorized operations.
	Locked       money.Decimal `json:"locked"`       // Amount reserved by Holds or other authorizations.
	Frozen       money.Decimal `json:"frozen"`       // Amount unavailable because of risk or operational freezes.
	Enabled      bool          `json:"enabled"`      // Whether the account is enabled for this App.
	Capabilities []Capability  `json:"capabilities"` // Operations authorized for this account key.
}
