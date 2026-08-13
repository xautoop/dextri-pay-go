package money

import "testing"

func TestValidation(t *testing.T) {
	for _, value := range []Decimal{"0", "100.00", "0.000001"} {
		if err := value.Validate(); err != nil {
			t.Errorf("%q: %v", value, err)
		}
	}
	for _, value := range []Decimal{"", "-1", "1e3", ".5", "01", "1."} {
		if err := value.Validate(); err == nil {
			t.Errorf("accepted %q", value)
		}
	}
	if err := Decimal("0").ValidatePositive(); err == nil {
		t.Fatal("zero accepted")
	}
}
