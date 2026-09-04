package escrow

import (
	"testing"

	"github.com/xautoop/dextri-pay-go/money"
)

func TestCommitRequestValidate(t *testing.T) {
	if err := (CommitRequest{ClientReferenceID: "order-1", HoldIDs: []string{"h1", "h2"}}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (CommitRequest{ClientReferenceID: "order-1", HoldIDs: []string{"h1", "h1"}}).Validate(); err == nil {
		t.Fatal("duplicate hold accepted")
	}
}

func TestSettlementRequestValidateAtomicConservation(t *testing.T) {
	valid := SettlementRequest{
		EscrowID: "e1", ClientReferenceID: "settle-1", Asset: "DXS",
		TotalAmount: "1000.00", TotalAmountAtomic: "1000000000",
		Allocations: []Allocation{
			{TargetType: TargetExternalUser, TargetID: "winner", Amount: money.Decimal("700.00"), AmountAtomic: "700000000"},
			{TargetType: TargetAppAccount, TargetID: "revenue", Amount: money.Decimal("200.00"), AmountAtomic: "200000000"},
			{TargetType: TargetAppAccount, TargetID: "pool", Amount: money.Decimal("50.00"), AmountAtomic: "50000000"},
			{TargetType: TargetAppAccount, TargetID: "burn", Amount: money.Decimal("50.00"), AmountAtomic: "50000000"},
		},
	}
	if err := valid.Validate(); err != nil {
		t.Fatal(err)
	}
	invalid := valid
	invalid.TotalAmountAtomic = "999999999"
	if err := invalid.Validate(); err == nil {
		t.Fatal("non-conserving allocations accepted")
	}
	invalid = valid
	invalid.Allocations = append([]Allocation(nil), valid.Allocations...)
	invalid.Allocations[3].TargetID = "pool"
	if err := invalid.Validate(); err == nil {
		t.Fatal("duplicate target accepted")
	}
}
