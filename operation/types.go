// Package operation defines durable Pay operation records and filters.
package operation

import (
	"time"

	"github.com/xautoop/dextri-pay-go/api"
	"github.com/xautoop/dextri-pay-go/money"
)

// Type identifies a durable Pay business workflow.
type Type string

const (
	TypeDeposit           Type = "deposit"
	TypeWithdrawal        Type = "withdrawal"
	TypeConversion        Type = "conversion"
	TypeDepositAndConvert Type = "deposit_and_convert"
	TypeUserBinding       Type = "user_binding"
)

// Status is the durable operation lifecycle state.
type Status string

const (
	StatusCreated            Status = "created"
	StatusAwaitingWallet     Status = "awaiting_wallet"
	StatusAwaitingUserAction Status = "awaiting_user_action"
	StatusProcessing         Status = "processing"
	StatusSucceeded          Status = "succeeded"
	StatusFailed             Status = "failed"
	StatusExpired            Status = "expired"
	StatusManualReview       Status = "manual_review"
)

// Operation is one App-visible durable business record.
type Operation struct {
	// OperationID is the Pay operation identifier.
	OperationID string `json:"operation_id"`
	// ClientReferenceID is the App-side business reference.
	ClientReferenceID string `json:"client_reference_id,omitempty"`
	// ExternalUserID is the App-side stable user identifier.
	ExternalUserID string `json:"external_user_id"`
	// Type identifies the business workflow.
	Type Type `json:"type"`
	// Status is the latest durable operation state.
	Status Status `json:"status"`
	// SourceAsset is the operation input asset.
	SourceAsset string `json:"source_asset,omitempty"`
	// TargetAsset is the operation output asset.
	TargetAsset string `json:"target_asset,omitempty"`
	// InputAmount is the human-readable accepted input amount.
	InputAmount money.Decimal `json:"input_amount,omitempty"`
	// OutputAmount is the human-readable final output amount.
	OutputAmount money.Decimal `json:"output_amount,omitempty"`
	// FailureCode is the stable machine-readable failure code.
	FailureCode string `json:"failure_code,omitempty"`
	// FailureReason is a safe human-readable failure description.
	FailureReason string `json:"failure_reason,omitempty"`
	// Metadata is the non-sensitive context supplied by the App.
	Metadata api.Metadata `json:"metadata,omitempty"`
	// CreatedAt is the operation creation time.
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is the latest persisted update time.
	UpdatedAt time.Time `json:"updated_at"`
}

// ListParams filters operations visible to the authenticated App.
type ListParams struct {
	// ExternalUserID optionally filters by App-side user identifier.
	ExternalUserID string
	// Type optionally filters by workflow type.
	Type Type
	// Status optionally filters by durable state.
	Status Status
	// Cursor continues a previous page.
	Cursor string
	// Limit requests the page size accepted by the API.
	Limit int
}

// Validate checks operation-list filters before an HTTP request is sent.
func (params ListParams) Validate() error {
	if params.Limit < 0 {
		return &api.ValidationError{Field: "limit", Message: "must not be negative"}
	}
	return nil
}

// List is one cursor-paginated operation page.
type List struct {
	// Items contains the operations in this page.
	Items []Operation `json:"items"`
	// NextCursor continues the listing when non-empty.
	NextCursor string `json:"next_cursor,omitempty"`
}
