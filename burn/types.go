// Package burn defines generic, auditable asset-destruction contracts.
package burn

import (
	"regexp"
	"strings"
	"time"

	"github.com/xautoop/dextri-pay-go/api"
	"github.com/xautoop/dextri-pay-go/money"
)

var atomicPattern = regexp.MustCompile(`^(0|[1-9][0-9]*)$`)

// Status is the authoritative lifecycle state of one burn request.
type Status string

const (
	StatusProcessing   Status = "processing"
	StatusSucceeded    Status = "succeeded"
	StatusFailed       Status = "failed"
	StatusManualReview Status = "manual_review"
)

// CreateRequest destroys an explicit amount from a registered App account.
// Pay resolves SourceAccountKey to its configured Vault account and verifies
// that the account has the burn capability; callers cannot submit an Owner or
// subaccount ID directly.
type CreateRequest struct {
	SourceAccountKey  string        `json:"source_account_key"`  // Registered App account key to debit.
	ClientReferenceID string        `json:"client_reference_id"` // App-side immutable burn reference.
	Asset             string        `json:"asset"`               // Asset code to destroy.
	Amount            money.Decimal `json:"amount"`              // Human-readable amount.
	AmountAtomic      string        `json:"amount_atomic"`       // Same amount in smallest units.
	Reason            string        `json:"reason"`              // Stable business reason.
	Metadata          api.Metadata  `json:"metadata,omitempty"`  // Non-sensitive audit context.
}

// Validate checks the public burn request before any network call.
func (request CreateRequest) Validate() error {
	for field, value := range map[string]string{
		"source_account_key":  request.SourceAccountKey,
		"client_reference_id": request.ClientReferenceID,
		"asset":               request.Asset,
		"reason":              request.Reason,
	} {
		if strings.TrimSpace(value) == "" {
			return &api.ValidationError{Field: field, Message: "is required"}
		}
	}
	if err := request.Amount.ValidatePositive(); err != nil {
		return &api.ValidationError{Field: "amount", Message: err.Error()}
	}
	if !atomicPattern.MatchString(request.AmountAtomic) || request.AmountAtomic == "0" {
		return &api.ValidationError{Field: "amount_atomic", Message: "must be a positive canonical integer"}
	}
	return nil
}

// Burn is the authoritative result of an idempotent destruction request.
type Burn struct {
	BurnID            string        `json:"burn_id"`                    // Pay burn identifier.
	OperationID       string        `json:"operation_id"`               // Durable operation identifier.
	SourceAccountKey  string        `json:"source_account_key"`         // Registered App account debited.
	ClientReferenceID string        `json:"client_reference_id"`        // App-side immutable reference.
	Asset             string        `json:"asset"`                      // Destroyed asset code.
	Amount            money.Decimal `json:"amount"`                     // Human-readable destroyed amount.
	AmountAtomic      string        `json:"amount_atomic"`              // Destroyed amount in smallest units.
	Reason            string        `json:"reason"`                     // Stable business reason.
	Status            Status        `json:"status"`                     // Latest authoritative status.
	TransactionHash   string        `json:"transaction_hash,omitempty"` // Final chain transaction hash.
	FailureCode       string        `json:"failure_code,omitempty"`     // Stable failure code.
	FailureReason     string        `json:"failure_reason,omitempty"`   // Safe failure description.
	Metadata          api.Metadata  `json:"metadata,omitempty"`         // Non-sensitive audit context.
	CreatedAt         time.Time     `json:"created_at"`                 // Request creation time.
	UpdatedAt         time.Time     `json:"updated_at"`                 // Latest authoritative update time.
}
