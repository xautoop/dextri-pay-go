// Package escrow defines generic multi-Hold commit and atomic allocation contracts.
package escrow

import (
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/xautoop/dextri-pay-go/api"
	"github.com/xautoop/dextri-pay-go/money"
)

var atomicPattern = regexp.MustCompile(`^(0|[1-9][0-9]*)$`)

// Status is the authoritative lifecycle state of committed escrow funds.
type Status string

const (
	StatusProcessing   Status = "processing"
	StatusCommitted    Status = "committed"
	StatusConsumed     Status = "consumed"
	StatusReleased     Status = "released"
	StatusFailed       Status = "failed"
	StatusManualReview Status = "manual_review"
)

// CommitRequest atomically commits one or more compatible Holds into an Escrow.
type CommitRequest struct {
	ClientReferenceID string       `json:"client_reference_id"` // App order-level business reference.
	HoldIDs           []string     `json:"hold_ids"`            // Unique Holds that must commit together.
	Metadata          api.Metadata `json:"metadata,omitempty"`  // Non-sensitive App context.
}

// Validate checks required identifiers and rejects duplicate Hold IDs.
func (request CommitRequest) Validate() error {
	if strings.TrimSpace(request.ClientReferenceID) == "" {
		return &api.ValidationError{Field: "client_reference_id", Message: "is required"}
	}
	if len(request.HoldIDs) == 0 {
		return &api.ValidationError{Field: "hold_ids", Message: "must not be empty"}
	}
	seen := make(map[string]struct{}, len(request.HoldIDs))
	for _, value := range request.HoldIDs {
		id := strings.TrimSpace(value)
		if id == "" {
			return &api.ValidationError{Field: "hold_ids", Message: "must not contain empty IDs"}
		}
		if _, exists := seen[id]; exists {
			return &api.ValidationError{Field: "hold_ids", Message: "must contain unique IDs"}
		}
		seen[id] = struct{}{}
	}
	return nil
}

// Escrow is the authoritative result of atomically committing compatible Holds.
type Escrow struct {
	EscrowID          string        `json:"escrow_id"`                // Pay Escrow identifier.
	OperationID       string        `json:"operation_id"`             // Durable commit operation identifier.
	ClientReferenceID string        `json:"client_reference_id"`      // App business reference.
	HoldIDs           []string      `json:"hold_ids"`                 // Holds committed into this Escrow.
	Asset             string        `json:"asset"`                    // Common asset of all committed Holds.
	Amount            money.Decimal `json:"amount"`                   // Total committed display amount.
	AmountAtomic      string        `json:"amount_atomic"`            // Total committed amount in smallest units.
	Status            Status        `json:"status"`                   // Latest authoritative Escrow status.
	FailureCode       string        `json:"failure_code,omitempty"`   // Stable failure code.
	FailureReason     string        `json:"failure_reason,omitempty"` // Safe failure description.
	Metadata          api.Metadata  `json:"metadata,omitempty"`       // Non-sensitive App context.
	CreatedAt         time.Time     `json:"created_at"`               // Escrow creation time.
	UpdatedAt         time.Time     `json:"updated_at"`               // Latest authoritative update time.
}

// TargetType identifies the kind of destination receiving an allocation.
type TargetType string

const (
	TargetExternalUser TargetType = "external_user"
	TargetAppAccount   TargetType = "app_account"
)

// Allocation is one explicit destination in an atomic Settlement.
type Allocation struct {
	TargetType   TargetType    `json:"target_type"`   // External user or registered App account.
	TargetID     string        `json:"target_id"`     // External user ID or App account key.
	Amount       money.Decimal `json:"amount"`        // Positive display amount.
	AmountAtomic string        `json:"amount_atomic"` // Same amount in smallest units.
}

// SettlementRequest atomically consumes an Escrow and distributes its full amount.
type SettlementRequest struct {
	EscrowID          string        `json:"escrow_id"`           // Committed Escrow to consume.
	ClientReferenceID string        `json:"client_reference_id"` // App settlement business reference.
	Asset             string        `json:"asset"`               // Settlement asset code.
	TotalAmount       money.Decimal `json:"total_amount"`        // Full display amount to distribute.
	TotalAmountAtomic string        `json:"total_amount_atomic"` // Full amount in smallest units.
	Allocations       []Allocation  `json:"allocations"`         // Explicit destinations whose atomic amounts must conserve the total.
	Metadata          api.Metadata  `json:"metadata,omitempty"`  // Non-sensitive evidence and audit context.
}

// Validate checks identifiers, destination types, decimal syntax and atomic conservation.
func (request SettlementRequest) Validate() error {
	for field, value := range map[string]string{
		"escrow_id":           request.EscrowID,
		"client_reference_id": request.ClientReferenceID,
		"asset":               request.Asset,
	} {
		if strings.TrimSpace(value) == "" {
			return &api.ValidationError{Field: field, Message: "is required"}
		}
	}
	if err := request.TotalAmount.ValidatePositive(); err != nil {
		return &api.ValidationError{Field: "total_amount", Message: err.Error()}
	}
	totalAtomic, ok := parsePositiveAtomic(request.TotalAmountAtomic)
	if !ok {
		return &api.ValidationError{Field: "total_amount_atomic", Message: "must be a positive canonical integer"}
	}
	if len(request.Allocations) == 0 {
		return &api.ValidationError{Field: "allocations", Message: "must not be empty"}
	}
	sum := new(big.Int)
	seen := make(map[string]struct{}, len(request.Allocations))
	for index, allocation := range request.Allocations {
		if allocation.TargetType != TargetExternalUser && allocation.TargetType != TargetAppAccount {
			return &api.ValidationError{Field: "allocations", Message: "contains an unsupported target_type"}
		}
		id := strings.TrimSpace(allocation.TargetID)
		if id == "" {
			return &api.ValidationError{Field: "allocations", Message: "contains an empty target_id"}
		}
		key := string(allocation.TargetType) + ":" + id
		if _, exists := seen[key]; exists {
			return &api.ValidationError{Field: "allocations", Message: "contains duplicate targets"}
		}
		seen[key] = struct{}{}
		if err := allocation.Amount.ValidatePositive(); err != nil {
			return &api.ValidationError{Field: "allocations", Message: "contains an invalid amount at index " + strconv.Itoa(index)}
		}
		atomic, valid := parsePositiveAtomic(allocation.AmountAtomic)
		if !valid {
			return &api.ValidationError{Field: "allocations", Message: "contains an invalid amount_atomic"}
		}
		sum.Add(sum, atomic)
	}
	if sum.Cmp(totalAtomic) != 0 {
		return &api.ValidationError{Field: "allocations", Message: "atomic amounts must equal total_amount_atomic"}
	}
	return nil
}

func parsePositiveAtomic(value string) (*big.Int, bool) {
	if !atomicPattern.MatchString(value) || value == "0" {
		return nil, false
	}
	parsed, ok := new(big.Int).SetString(value, 10)
	return parsed, ok
}

// Settlement is the authoritative result of an atomic Escrow distribution.
type Settlement struct {
	SettlementID      string        `json:"settlement_id"`            // Pay Settlement identifier.
	OperationID       string        `json:"operation_id"`             // Durable settlement operation identifier.
	EscrowID          string        `json:"escrow_id"`                // Consumed Escrow identifier.
	ClientReferenceID string        `json:"client_reference_id"`      // App business reference.
	Asset             string        `json:"asset"`                    // Settled asset code.
	TotalAmount       money.Decimal `json:"total_amount"`             // Distributed display amount.
	TotalAmountAtomic string        `json:"total_amount_atomic"`      // Distributed amount in smallest units.
	Allocations       []Allocation  `json:"allocations"`              // Authoritative allocation result.
	Status            Status        `json:"status"`                   // Latest authoritative Settlement status.
	FailureCode       string        `json:"failure_code,omitempty"`   // Stable failure code.
	FailureReason     string        `json:"failure_reason,omitempty"` // Safe failure description.
	Metadata          api.Metadata  `json:"metadata,omitempty"`       // Non-sensitive App context.
	CreatedAt         time.Time     `json:"created_at"`               // Settlement creation time.
	UpdatedAt         time.Time     `json:"updated_at"`               // Latest authoritative update time.
}
