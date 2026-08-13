// Package channels defines channel-discovery API contracts.
package channels

import (
	"github.com/xautoop/dextri-pay-go/api"
	"github.com/xautoop/dextri-pay-go/money"
)

// Flow identifies the direction of a funding channel.
type Flow string

const (
	FlowDeposit    Flow = "deposit"
	FlowWithdrawal Flow = "withdrawal"
)

// ListParams filters channels available to the authenticated App.
type ListParams struct {
	// Flow optionally limits results to deposit or withdrawal channels.
	Flow Flow
	// SourceAsset optionally limits results to one source asset.
	SourceAsset string
}

// Channel is one currently authorized and healthy funding route.
type Channel struct {
	// ID is the stable channel identifier passed to Hosted Checkout.
	ID string `json:"id"`
	// Flow is deposit or withdrawal.
	Flow Flow `json:"flow"`
	// SourceChainNamespace identifies the source chain family.
	SourceChainNamespace string `json:"source_chain_namespace"`
	// SourceChainID identifies the source network.
	SourceChainID string `json:"source_chain_id"`
	// SourceAsset is the asset accepted from the user.
	SourceAsset string `json:"source_asset"`
	// DestinationChainID identifies the settlement network.
	DestinationChainID string `json:"destination_chain_id"`
	// DestinationAsset is the asset produced by the route.
	DestinationAsset string `json:"destination_asset"`
	// RouteFamily identifies the execution mechanism.
	RouteFamily string `json:"route_family"`
	// Provider identifies the route provider.
	Provider string `json:"provider"`
	// MinAmount is the minimum human-readable route amount.
	MinAmount money.Decimal `json:"min_amount,omitempty"`
	// MaxAmount is the maximum human-readable route amount.
	MaxAmount money.Decimal `json:"max_amount,omitempty"`
	// Healthy reports whether the route may be selected.
	Healthy bool `json:"healthy"`
	// UnavailableReason explains why a route is unavailable.
	UnavailableReason string `json:"unavailable_reason,omitempty"`
	// Metadata contains forward-compatible, non-sensitive route details.
	Metadata api.Metadata `json:"metadata,omitempty"`
}
