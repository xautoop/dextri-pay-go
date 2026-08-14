// Package payout defines optional App-funded disbursement contracts. A payout
// can represent a reward, rebate, compensation, or another App business use.
package payout

import (
	"strings"

	"github.com/xautoop/dextri-pay-go/api"
	"github.com/xautoop/dextri-pay-go/money"
	"github.com/xautoop/dextri-pay-go/operation"
)

type CreateRequest struct {
	ExternalUserID    string        `json:"external_user_id"`
	ClientReferenceID string        `json:"client_reference_id"`
	Asset             string        `json:"asset"`
	Amount            money.Decimal `json:"amount"`
	Reason            string        `json:"reason"`
	Metadata          api.Metadata  `json:"metadata,omitempty"`
}

func (request CreateRequest) Validate() error {
	for field, value := range map[string]string{"external_user_id": request.ExternalUserID, "client_reference_id": request.ClientReferenceID, "asset": request.Asset, "reason": request.Reason} {
		if strings.TrimSpace(value) == "" {
			return &api.ValidationError{Field: field, Message: "is required"}
		}
	}
	if err := request.Amount.ValidatePositive(); err != nil {
		return &api.ValidationError{Field: "amount", Message: err.Error()}
	}
	return nil
}

type Payout = operation.Operation
