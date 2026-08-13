package users

import "testing"

func TestCreateBindingSessionRequestValidate(t *testing.T) {
	valid := CreateBindingSessionRequest{ExternalUserID: "user", WalletFamily: WalletFamilyEVM}
	if err := valid.Validate(); err != nil {
		t.Fatal(err)
	}
	for _, request := range []CreateBindingSessionRequest{
		{WalletFamily: WalletFamilyEVM},
		{ExternalUserID: "user", WalletFamily: "unknown"},
	} {
		if err := request.Validate(); err == nil {
			t.Fatalf("invalid request accepted: %#v", request)
		}
	}
}
