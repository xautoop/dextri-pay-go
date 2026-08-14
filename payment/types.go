// Package payment defines App Commerce payment and refund contracts.
package payment

import (
	"strings"

	"github.com/xautoop/dextri-pay-go/api"
	"github.com/xautoop/dextri-pay-go/operation"
)

type Payment = operation.Operation

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
