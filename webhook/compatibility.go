package webhook

import (
	"encoding/json"

	"github.com/xautoop/dextri-pay-go/money"
	"github.com/xautoop/dextri-pay-go/operation"
)

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
