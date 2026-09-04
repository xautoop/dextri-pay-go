// Package hold defines generic App-authorized balance reservation contracts.
package hold

import (
	"strings"
	"time"

	"github.com/xautoop/dextri-pay-go/api"
	"github.com/xautoop/dextri-pay-go/money"
)

// Status is the authoritative lifecycle state of a Hold.
type Status string

const (
	StatusProcessing     Status = "processing"
	StatusHeld           Status = "held"
	StatusCommitPending  Status = "commit_pending"
	StatusCommitted      Status = "committed"
	StatusReleasePending Status = "release_pending"
	StatusReleased       Status = "released"
	StatusConsumePending Status = "consume_pending"
	StatusConsumed       Status = "consumed"
	StatusFailed         Status = "failed"
	StatusManualReview   Status = "manual_review"
)

// CreateRequest creates one expiring reservation against a user's available balance.
type CreateRequest struct {
	ExternalUserID    string        `json:"external_user_id"`    // Stable user identifier inside the authenticated App.
	ClientReferenceID string        `json:"client_reference_id"` // App business reference used for reconciliation.
	Asset             string        `json:"asset"`               // Asset code.
	Amount            money.Decimal `json:"amount"`              // Positive display amount encoded as a decimal string.
	Purpose           string        `json:"purpose"`             // App-neutral business purpose.
	ExpiresAt         time.Time     `json:"expires_at"`          // Time after which an uncommitted Hold enters release processing.
	Metadata          api.Metadata  `json:"metadata,omitempty"`  // Non-sensitive App context.
}

// Validate checks a Hold request before any HTTP call is made.
func (request CreateRequest) Validate() error {
	for field, value := range map[string]string{
		"external_user_id":    request.ExternalUserID,
		"client_reference_id": request.ClientReferenceID,
		"asset":               request.Asset,
		"purpose":             request.Purpose,
	} {
		if strings.TrimSpace(value) == "" {
			return &api.ValidationError{Field: field, Message: "is required"}
		}
	}
	if err := request.Amount.ValidatePositive(); err != nil {
		return &api.ValidationError{Field: "amount", Message: err.Error()}
	}
	if request.ExpiresAt.IsZero() {
		return &api.ValidationError{Field: "expires_at", Message: "is required"}
	}
	return nil
}

// ReleaseRequest requests release of an unconsumed Hold.
type ReleaseRequest struct {
	Reason   string       `json:"reason"`             // Stable release reason supplied by the App.
	Metadata api.Metadata `json:"metadata,omitempty"` // Non-sensitive audit context.
}

// Validate checks a release request before any HTTP call is made.
func (request ReleaseRequest) Validate() error {
	if strings.TrimSpace(request.Reason) == "" {
		return &api.ValidationError{Field: "reason", Message: "is required"}
	}
	return nil
}

// Hold is the authoritative Pay representation of a balance reservation.
type Hold struct {
	HoldID            string        `json:"hold_id"`                  // Pay Hold identifier.
	OperationID       string        `json:"operation_id"`             // Durable operation identifier for status recovery.
	ExternalUserID    string        `json:"external_user_id"`         // App user whose balance is reserved.
	ClientReferenceID string        `json:"client_reference_id"`      // App business reference.
	Asset             string        `json:"asset"`                    // Reserved asset code.
	Amount            money.Decimal `json:"amount"`                   // Reserved display amount.
	Purpose           string        `json:"purpose"`                  // App-neutral business purpose.
	Status            Status        `json:"status"`                   // Latest authoritative Hold status.
	ExpiresAt         *time.Time    `json:"expires_at,omitempty"`     // Expiry while the Hold remains uncommitted.
	FailureCode       string        `json:"failure_code,omitempty"`   // Stable failure code when processing failed.
	FailureReason     string        `json:"failure_reason,omitempty"` // Safe failure description.
	Metadata          api.Metadata  `json:"metadata,omitempty"`       // Non-sensitive App context.
	CreatedAt         time.Time     `json:"created_at"`               // Hold creation time.
	UpdatedAt         time.Time     `json:"updated_at"`               // Latest authoritative update time.
}
