package tron

import "testing"

func TestWithdrawalReviewRequestsValidate(t *testing.T) {
	tests := []struct {
		name    string
		request interface{ Validate() error }
		wantErr bool
	}{
		{name: "approve", request: ReviewRequest{Actor: "pvsz:admin:1"}},
		{name: "approve missing actor", request: ReviewRequest{}, wantErr: true},
		{name: "reject", request: RejectRequest{Actor: "pvsz:admin:1", Reason: "risk"}},
		{name: "reject missing reason", request: RejectRequest{Actor: "pvsz:admin:1"}, wantErr: true},
		{name: "transfer", request: SubmitTransferRequest{Actor: "pvsz:admin:1", TxHash: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"}},
		{name: "transfer invalid hash", request: SubmitTransferRequest{Actor: "pvsz:admin:1", TxHash: "not-a-tx"}, wantErr: true},
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
