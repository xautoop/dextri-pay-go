package operation

import "testing"

func TestListParamsValidate(t *testing.T) {
	if err := (ListParams{Limit: 100}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (ListParams{Limit: -1}).Validate(); err == nil {
		t.Fatal("negative limit accepted")
	}
}
