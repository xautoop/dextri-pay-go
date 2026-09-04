package money

import "testing"

func TestSplitAtomicKeepsTwoDecimalOutputsAndConservesTotal(t *testing.T) {
	result, err := SplitAtomic("100010000", 6, 2, []BasisPointShare{
		{Key: "winner", BPS: 7000},
		{Key: "platform", BPS: 2000},
		{Key: "weekly_pool", BPS: 500},
		{Key: "burn_pending", BPS: 500},
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"70010000", "20000000", "5000000", "5000000"}
	for index := range want {
		if result[index].AmountAtomic != want[index] {
			t.Fatalf("result[%d]=%s want=%s", index, result[index].AmountAtomic, want[index])
		}
	}
}

func TestSplitAtomicRejectsTotalBeyondOutputPrecision(t *testing.T) {
	_, err := SplitAtomic("100000001", 6, 2, []BasisPointShare{{Key: "winner", BPS: 10000}})
	if err == nil {
		t.Fatal("expected output precision validation error")
	}
}

func TestSplitAtomicUsesInputOrderForEqualRemainders(t *testing.T) {
	result, err := SplitAtomic("100", 2, 2, []BasisPointShare{
		{Key: "first", BPS: 3334}, {Key: "second", BPS: 3333}, {Key: "third", BPS: 3333},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result[0].AmountAtomic != "34" || result[1].AmountAtomic != "33" || result[2].AmountAtomic != "33" {
		t.Fatalf("unexpected deterministic allocation: %+v", result)
	}
}
