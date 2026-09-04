package hold

import (
	"testing"
	"time"

	"github.com/xautoop/dextri-pay-go/money"
)

func TestCreateRequestValidate(t *testing.T) {
	valid := CreateRequest{
		ExternalUserID: "user-1", ClientReferenceID: "order-1", Asset: "DXS",
		Amount: money.Decimal("500.00"), Purpose: "game_stake", ExpiresAt: time.Now().Add(time.Hour),
	}
	if err := valid.Validate(); err != nil {
		t.Fatal(err)
	}
	invalid := valid
	invalid.Amount = "0"
	if err := invalid.Validate(); err == nil {
		t.Fatal("zero amount accepted")
	}
	invalid = valid
	invalid.ExpiresAt = time.Time{}
	if err := invalid.Validate(); err == nil {
		t.Fatal("empty expiry accepted")
	}
}

func TestReleaseRequestValidate(t *testing.T) {
	if err := (ReleaseRequest{Reason: "expired"}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (ReleaseRequest{}).Validate(); err == nil {
		t.Fatal("empty reason accepted")
	}
}
