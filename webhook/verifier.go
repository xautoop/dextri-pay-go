// Package webhook verifies and decodes signed Pay webhook deliveries.
package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/xautoop/dextri-pay-go/money"
	"github.com/xautoop/dextri-pay-go/operation"
)

const (
	SignatureScheme  = "DEXTRI-PAY-WEBHOOK-HMAC-SHA256"
	HeaderDeliveryID = "X-DEXTRI-PAY-WEBHOOK-ID"
	HeaderTimestamp  = "X-DEXTRI-PAY-WEBHOOK-TS"
	HeaderSignature  = "X-DEXTRI-PAY-WEBHOOK-SIGN"
	DefaultTolerance = 5 * time.Minute
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
}

// UnmarshalJSON accepts both the current public Operation contract and the
// legacy Outbox field names used by already queued webhook deliveries.
func (data *EventData) UnmarshalJSON(raw []byte) error {
	var envelope struct {
		Operation json.RawMessage `json:"operation"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return err
	}
	if len(envelope.Operation) == 0 || string(envelope.Operation) == "null" {
		return nil
	}

	var current operation.Operation
	if err := json.Unmarshal(envelope.Operation, &current); err != nil {
		return err
	}
	var legacy struct {
		DestinationAsset string        `json:"destination_asset"`
		Amount           money.Decimal `json:"amount"`
	}
	if err := json.Unmarshal(envelope.Operation, &legacy); err != nil {
		return err
	}
	if current.TargetAsset == "" {
		current.TargetAsset = legacy.DestinationAsset
	}
	if current.InputAmount == "" {
		current.InputAmount = legacy.Amount
	}
	data.Operation = current
	return nil
}

// Delivery contains verified transport metadata and the decoded event.
type Delivery struct {
	// DeliveryID is unique to one delivery attempt stream.
	DeliveryID string
	// SignedAt is the timestamp covered by the webhook signature.
	SignedAt time.Time
	// Event is the verified and decoded payload.
	Event Event
}

// Verifier verifies webhook headers, timestamp tolerance and payload signature.
type Verifier struct {
	// Secret is the webhook endpoint secret shown once by Admin.
	Secret string
	// Tolerance overrides DefaultTolerance when non-zero.
	Tolerance time.Duration
	// Now optionally supplies a deterministic clock for testing.
	Now func() time.Time
}

// Verify authenticates delivery headers, checks freshness and decodes the event.
func (verifier Verifier) Verify(headers http.Header, body []byte) (*Delivery, error) {
	if verifier.Secret == "" {
		return nil, errors.New("webhook secret is required")
	}
	id := strings.TrimSpace(headers.Get(HeaderDeliveryID))
	timestampRaw := strings.TrimSpace(headers.Get(HeaderTimestamp))
	provided := strings.ToLower(strings.TrimSpace(headers.Get(HeaderSignature)))
	if id == "" || timestampRaw == "" || provided == "" {
		return nil, errors.New("webhook signature headers are incomplete")
	}
	timestampMS, err := strconv.ParseInt(timestampRaw, 10, 64)
	if err != nil {
		return nil, errors.New("webhook timestamp is invalid")
	}
	signedAt := time.UnixMilli(timestampMS)
	now := time.Now()
	if verifier.Now != nil {
		now = verifier.Now()
	}
	tolerance := verifier.Tolerance
	if tolerance == 0 {
		tolerance = DefaultTolerance
	}
	delta := now.Sub(signedAt)
	if delta < 0 {
		delta = -delta
	}
	if tolerance < 0 || delta > tolerance {
		return nil, errors.New("webhook timestamp is outside the allowed window")
	}
	expectedBytes, _ := hex.DecodeString(sign(verifier.Secret, id, timestampRaw, body))
	providedBytes, decodeErr := hex.DecodeString(provided)
	if decodeErr != nil || !hmac.Equal(expectedBytes, providedBytes) {
		return nil, errors.New("webhook signature verification failed")
	}
	var event Event
	if err := json.Unmarshal(body, &event); err != nil {
		return nil, errors.New("webhook payload is invalid JSON")
	}
	if strings.TrimSpace(event.ID) == "" || strings.TrimSpace(string(event.Type)) == "" {
		return nil, errors.New("webhook payload is missing id or type")
	}
	return &Delivery{DeliveryID: id, SignedAt: signedAt, Event: event}, nil
}

func sign(secret, id, timestamp string, body []byte) string {
	hash := sha256.Sum256(body)
	payload := strings.Join([]string{SignatureScheme, strings.TrimSpace(id), strings.TrimSpace(timestamp), hex.EncodeToString(hash[:])}, "\n")
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}
