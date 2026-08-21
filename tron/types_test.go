package tron

import (
	"encoding/json"
	"testing"
)

func TestWithdrawalDecodesFeeSnapshot(t *testing.T) {
	var withdrawal Withdrawal
	err := json.Unmarshal([]byte(`{
		"withdrawal_id":"paywd_1",
		"amount":"90",
		"amount_atomic":"90000000",
		"withdrawal_amount":"100",
		"fee_amount":"10",
		"receive_amount":"90",
		"fee_rate_bps":1000
	}`), &withdrawal)
	if err != nil {
		t.Fatalf("decode withdrawal: %v", err)
	}
	if withdrawal.WithdrawalAmount.String() != "100" || withdrawal.FeeAmount.String() != "10" || withdrawal.ReceiveAmount.String() != "90" || withdrawal.FeeRateBPS != 1000 {
		t.Fatalf("fee snapshot = %+v", withdrawal)
	}
}

func TestWithdrawalReviewRequestsValidate(t *testing.T) {
	tests := []struct {
		name    string
		request interface{ Validate() error }
		wantErr bool
	}{
		{name: "approve", request: ApproveRequest{Actor: "pvsz:admin:1", SourceAddress: "T9yD14Nj9j7xQy7w7r1aQmJxK3Y7JxYv9x"}},
		{name: "approve missing actor", request: ApproveRequest{SourceAddress: "T9yD14Nj9j7xQy7w7r1aQmJxK3Y7JxYv9x"}, wantErr: true},
		{name: "approve missing source", request: ApproveRequest{Actor: "pvsz:admin:1"}, wantErr: true},
		{name: "reject", request: RejectRequest{Actor: "pvsz:admin:1", Reason: "risk"}},
		{name: "reject missing reason", request: RejectRequest{Actor: "pvsz:admin:1"}, wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.request.Validate()
			if (err != nil) != test.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}
