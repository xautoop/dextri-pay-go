package webhook

import (
	"time"

	"github.com/xautoop/dextri-pay-go/escrow"
	"github.com/xautoop/dextri-pay-go/hold"
	"github.com/xautoop/dextri-pay-go/operation"
)

// EventType identifies a Pay webhook lifecycle event.
type EventType string

const (
	EventOperationCreated        EventType = "operation.created"
	EventOperationAwaitingWallet EventType = "operation.awaiting_wallet"
	EventOperationAwaitingAction EventType = "operation.awaiting_user_action"
	EventOperationProcessing     EventType = "operation.processing"
	EventOperationSucceeded      EventType = "operation.succeeded"
	EventOperationFailed         EventType = "operation.failed"
	EventOperationExpired        EventType = "operation.expired"
	EventOperationManualReview   EventType = "operation.manual_review"
	EventPaymentAuthorized       EventType = "payment.authorized"
	EventPaymentSucceeded        EventType = "payment.succeeded"
	EventPaymentCancelled        EventType = "payment.cancelled"
	EventRefundSucceeded         EventType = "refund.succeeded"
	EventRefundFailed            EventType = "refund.failed"
	EventPayoutSucceeded         EventType = "payout.succeeded"
	EventPayoutFailed            EventType = "payout.failed"
	EventHoldHeld                EventType = "hold.held"
	EventHoldFailed              EventType = "hold.failed"
	EventHoldReleased            EventType = "hold.released"
	EventEscrowCommitted         EventType = "escrow.committed"
	EventEscrowFailed            EventType = "escrow.failed"
	EventSettlementSucceeded     EventType = "escrow_settlement.succeeded"
	EventSettlementFailed        EventType = "escrow_settlement.failed"
)

// Event is the signed webhook payload delivered by Dextri Pay.
type Event struct {
	// ID is the stable event identifier used for consumer deduplication.
	ID string `json:"id"`
	// Type identifies the operation lifecycle transition.
	Type EventType `json:"type"`
	// CreatedAt is the event creation time.
	CreatedAt time.Time `json:"created_at"`
	// Data contains the latest durable resource snapshot.
	Data EventData `json:"data"`
}

// EventData contains resources associated with an event.
type EventData struct {
	// Operation is the latest App-visible operation snapshot.
	Operation operation.Operation `json:"operation"`
	// Hold is present for Hold creation and release events.
	Hold *hold.Hold `json:"hold,omitempty"`
	// Escrow is present for multi-Hold commit events.
	Escrow *escrow.Escrow `json:"escrow,omitempty"`
	// Settlement is present for atomic allocation events.
	Settlement *escrow.Settlement `json:"settlement,omitempty"`
}
