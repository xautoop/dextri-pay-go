package burn

import "testing"

func TestCreateRequestValidate(t *testing.T) {
	valid := CreateRequest{
		SourceAccountKey: "burn_pending", ClientReferenceID: "burn-week-1",
		Asset: "DXS", Amount: "50.00", AmountAtomic: "50000000", Reason: "scheduled_burn",
	}
	if err := valid.Validate(); err != nil {
		t.Fatal(err)
	}

	tests := []CreateRequest{
		{},
		{SourceAccountKey: "burn_pending", ClientReferenceID: "burn-week-1", Asset: "DXS", Amount: "50.00", AmountAtomic: "050000000", Reason: "scheduled_burn"},
		{SourceAccountKey: "burn_pending", ClientReferenceID: "burn-week-1", Asset: "DXS", Amount: "0", AmountAtomic: "1", Reason: "scheduled_burn"},
	}
	for _, request := range tests {
		if err := request.Validate(); err == nil {
			t.Fatalf("expected validation error for %#v", request)
		}
	}
}
