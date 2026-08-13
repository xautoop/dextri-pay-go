package channels

import "testing"

func TestListParamsValidate(t *testing.T) {
	for _, params := range []ListParams{{}, {Flow: FlowDeposit}, {Flow: FlowWithdrawal}} {
		if err := params.Validate(); err != nil {
			t.Fatalf("Validate(%#v) error = %v", params, err)
		}
	}
	if err := (ListParams{Flow: "unknown"}).Validate(); err == nil {
		t.Fatal("unknown flow accepted")
	}
}
