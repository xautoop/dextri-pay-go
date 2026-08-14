// Package payment defines App Commerce payment and refund contracts.
package payment

import (
	"strings"

	"github.com/xautoop/dextri-pay-go/api"
	"github.com/xautoop/dextri-pay-go/money"
	"github.com/xautoop/dextri-pay-go/operation"
)

type Payment = operation.Operation

// CreateRequest creates a direct App-custodial payment. Amount is a display
// decimal; Pay resolves and validates atomic precision from the chain registry.
type CreateRequest struct {
	ExternalUserID    string        `json:"external_user_id"`
	ClientReferenceID string        `json:"client_reference_id"`
	Asset             string        `json:"asset"`
	Amount            money.Decimal `json:"amount"`
	Metadata          api.Metadata  `json:"metadata,omitempty"`
}

// Validate checks required business identifiers before sending the request.
func (request CreateRequest) Validate() error {
	if strings.TrimSpace(request.ExternalUserID) == "" {
		return &api.ValidationError{Field: "external_user_id", Message: "is required"}
	}
	if strings.TrimSpace(request.ClientReferenceID) == "" {
		return &api.ValidationError{Field: "client_reference_id", Message: "is required"}
	}
	if strings.TrimSpace(request.Asset) == "" {
		return &api.ValidationError{Field: "asset", Message: "is required"}
	}
	if err := request.Amount.ValidatePositive(); err != nil {
		return &api.ValidationError{Field: "amount", Message: err.Error()}
	}
	return nil
}

// CreateResult is returned immediately after Pay durably reserves player funds.
// A payment_pending operation is complete only after later status reconciliation.
type CreateResult struct {
	Operation         operation.Operation `json:"operation"`
	SignatureRequired bool                `json:"signature_required"`
	InteractionType   string              `json:"interaction_type"`
}

// RefundRequest requests the v0.1 full-refund operation.
type RefundRequest struct {
	RefundID string `json:"refund_id"`
	Reason   string `json:"reason"`
}

func (request RefundRequest) Validate() error {
	if strings.TrimSpace(request.RefundID) == "" {
		return &api.ValidationError{Field: "refund_id", Message: "is required"}
	}
	if strings.TrimSpace(request.Reason) == "" {
		return &api.ValidationError{Field: "reason", Message: "is required"}
	}
	return nil
}

type Refund = operation.Operation
