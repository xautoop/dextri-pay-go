package tron

import "testing"

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
